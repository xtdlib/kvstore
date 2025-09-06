package kvstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var DRIVER = "sqlite"

var (
	sharedDB *sql.DB
	dbOnce   sync.Once
)

type KV[T1 any, T2 any] struct {
	db       *sql.DB
	table    string
	watchers *watcherRegistry[T1, T2]
}

type WatchEvent[T1 any, T2 any] struct {
	Type     WatchEventType
	Key      T1
	Value    T2
	OldValue T2
}

type WatchEventType int

const (
	WatchEventSet WatchEventType = iota
	WatchEventDelete
)

type watcher[T1 any, T2 any] struct {
	id       string
	key      *T1
	prefix   *string
	ch       chan WatchEvent[T1, T2]
	stopCh   chan struct{}
	stopped  bool
	stopOnce sync.Once
}

type watcherRegistry[T1 any, T2 any] struct {
	mu       sync.RWMutex
	watchers map[string]*watcher[T1, T2]
	store    *KV[T1, T2]
}

type CancelFunc func()

func getSharedDB() (*sql.DB, error) {
	var err error
	dbOnce.Do(func() {
		var cacheDir string

		// Get XDG cache directory
		cacheDir = os.Getenv("XDG_CACHE_HOME")
		if cacheDir == "" {
			home, homeErr := os.UserHomeDir()
			if homeErr != nil {
				// Fall back to /var/cache if home directory is not available
				cacheDir = "/var/cache"
			} else {
				cacheDir = filepath.Join(home, ".cache")
			}
		}

		// Create directory for this executable
		execName := filepath.Base(os.Args[0])
		dbDir := filepath.Join(cacheDir, execName)
		if mkdirErr := os.MkdirAll(dbDir, 0755); mkdirErr != nil {
			err = mkdirErr
			return
		}

		// Open the shared database with query parameters for better concurrency
		dbPath := filepath.Join(dbDir, execName+".db")
		// Add busy_timeout and other parameters directly in the connection string
		// connStr := fmt.Sprintf("file:%s?_timefmt=rfc3339", dbPath)
		connStr := fmt.Sprintf("file:%s", dbPath)
		log.Println(connStr)
		sharedDB, err = sql.Open(DRIVER, connStr)
		if err != nil {
			return
		}

		// Enable WAL mode and set pragmas for better concurrency
		// WAL allows concurrent reads and one writer
		pragmas := []string{
			"PRAGMA journal_mode=WAL",   // Enable Write-Ahead Logging
			"PRAGMA synchronous=NORMAL", // Good balance of safety and speed
			"PRAGMA busy_timeout=10000", // Wait up to 10 seconds when database is locked
			"PRAGMA cache_size=-32000",  // 32MB cache
		}

		for _, pragma := range pragmas {
			if _, pragmaErr := sharedDB.Exec(pragma); pragmaErr != nil {
				err = pragmaErr
				return
			}
		}

		// Set connection pool settings for better concurrency
		sharedDB.SetMaxOpenConns(25) // Allow multiple readers
		sharedDB.SetMaxIdleConns(25)
		sharedDB.SetConnMaxLifetime(5 * time.Minute)
	})
	return sharedDB, err
}

func NewAt[T1 any, T2 any](dbPath string, name string) (*KV[T1, T2], error) {
	connStr := fmt.Sprintf("%s?_busy_timeout=10000&_journal=WAL&_sync=NORMAL", dbPath)
	db, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, err
	}

	store := &KV[T1, T2]{
		db:    db,
		table: name,
	}

	store.watchers = &watcherRegistry[T1, T2]{
		watchers: make(map[string]*watcher[T1, T2]),
		store:    store,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Create table with sanitized name
	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (key PRIMARY KEY, value)", store.table)
	_, err = store.db.ExecContext(ctx, createSQL)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func New[T1 any, T2 any](name string) *KV[T1, T2] {
	db, err := getSharedDB()
	if err != nil {
		panic(err)
	}

	store := &KV[T1, T2]{
		db:    db,
		table: name,
	}

	store.watchers = &watcherRegistry[T1, T2]{
		watchers: make(map[string]*watcher[T1, T2]),
		store:    store,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Create table with sanitized name
	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (key PRIMARY KEY, value)", store.table)
	_, err = store.db.ExecContext(ctx, createSQL)
	if err != nil {
		panic(err)
	}

	return store
}

func (s *KV[T1, T2]) TrySet(key T1, value T2) (out T2, err error) {
	// Get old value for watch events
	oldValue, hadOldValue := s.getOldValue(key)

	// Serialize the value to JSON
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return out, fmt.Errorf("failed to marshal value: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sql := fmt.Sprintf("INSERT OR REPLACE INTO %s (key, value) VALUES (?, ?)", s.table)
	_, err = s.db.ExecContext(ctx, sql, key, string(valueBytes))
	if err != nil {
		return out, err
	}

	// Notify watchers
	if s.watchers != nil {
		event := WatchEvent[T1, T2]{
			Type:  WatchEventSet,
			Key:   key,
			Value: value,
		}
		if hadOldValue {
			event.OldValue = oldValue
		}
		s.watchers.notify(key, event)
	}

	return out, nil
}

func (s *KV[T1, T2]) TryGet(key T1) (T2, error) {
	var v T2
	var valueStr string
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sql := fmt.Sprintf("SELECT value FROM %s WHERE key = ?", s.table)
	err := s.db.QueryRowContext(ctx, sql, key).Scan(&valueStr)
	if err != nil {
		return v, err
	}
	
	// Deserialize from JSON
	err = json.Unmarshal([]byte(valueStr), &v)
	if err != nil {
		return v, fmt.Errorf("failed to unmarshal value: %w", err)
	}
	
	return v, nil
}

func (s *KV[T1, T2]) TryHas(key T1) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE key = ? LIMIT 1", s.table)
	var exists int
	err := s.db.QueryRowContext(ctx, query, key).Scan(&exists)
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return false, err
}

func (s *KV[T1, T2]) TryDelete(key T1) error {
	// Get old value for watch events
	oldValue, hadOldValue := s.getOldValue(key)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sql := fmt.Sprintf("DELETE FROM %s WHERE key = ?", s.table)
	result, err := s.db.ExecContext(ctx, sql, key)
	if err != nil {
		return err
	}

	// Only notify if something was actually deleted
	if s.watchers != nil && hadOldValue {
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			event := WatchEvent[T1, T2]{
				Type:     WatchEventDelete,
				Key:      key,
				OldValue: oldValue,
			}
			s.watchers.notify(key, event)
		}
	}

	return nil
}

func (s *KV[T1, T2]) TryForEach(fn func(key T1, value T2)) error {
	ctx := context.Background()
	sql := fmt.Sprintf("SELECT key, value FROM %s", s.table)
	rows, err := s.db.QueryContext(ctx, sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var k T1
		var v T2
		var valueStr string
		if err := rows.Scan(&k, &valueStr); err != nil {
			return err
		}
		
		// Deserialize from JSON
		if err := json.Unmarshal([]byte(valueStr), &v); err != nil {
			return fmt.Errorf("failed to unmarshal value: %w", err)
		}
		
		fn(k, v)
	}
	return rows.Err()
}

func (s *KV[T1, T2]) TryClear() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sql := fmt.Sprintf("DELETE FROM %s", s.table)
	_, err := s.db.ExecContext(ctx, sql)

	// Clear notifies all watchers with delete events
	// For simplicity, we're not sending individual delete events for each key

	return err
}

// Watch monitors changes to a specific key
func (s *KV[T1, T2]) Watch(key T1) (<-chan WatchEvent[T1, T2], CancelFunc) {
	ch := make(chan WatchEvent[T1, T2], 10)

	w := &watcher[T1, T2]{
		id:     fmt.Sprintf("%v_%d", key, time.Now().UnixNano()),
		key:    &key,
		ch:     ch,
		stopCh: make(chan struct{}),
	}

	s.watchers.mu.Lock()
	s.watchers.watchers[w.id] = w
	s.watchers.mu.Unlock()

	return ch, func() {
		w.stop()
		s.watchers.mu.Lock()
		delete(s.watchers.watchers, w.id)
		s.watchers.mu.Unlock()
		close(ch)
	}
}

// // WatchPrefix monitors changes to keys with a specific prefix
// func (s *KV[T1, T2]) WatchPrefix(prefix string) (<-chan WatchEvent[T1, T2], CancelFunc) {
// 	ch := make(chan WatchEvent[T1, T2], 10)
//
// 	w := &watcher[T1, T2]{
// 		id:     fmt.Sprintf("prefix_%s_%d", prefix, time.Now().UnixNano()),
// 		prefix: &prefix,
// 		ch:     ch,
// 		stopCh: make(chan struct{}),
// 	}
//
// 	s.watchers.mu.Lock()
// 	s.watchers.watchers[w.id] = w
// 	s.watchers.mu.Unlock()
//
// 	return ch, func() {
// 		w.stop()
// 		s.watchers.mu.Lock()
// 		delete(s.watchers.watchers, w.id)
// 		s.watchers.mu.Unlock()
// 		close(ch)
// 	}
// }

func (s *KV[T1, T2]) Set(key T1, value T2) T2 {
	out, err := s.TrySet(key, value)
	if err != nil {
		panic(err)
	}
	return out
}

func (s *KV[T1, T2]) Get(key T1) T2 {
	val, err := s.TryGet(key)
	if err != nil {
		panic(err)
	}
	return val
}

func (s *KV[T1, T2]) GetOr(key T1, defaultValue T2) T2 {
	val, err := s.TryGet(key)
	if errors.Is(err, sql.ErrNoRows) {
		return defaultValue
	} else if err != nil {
		panic(err)
	}
	return val
}

func (s *KV[T1, T2]) Has(key T1) bool {
	exists, err := s.TryHas(key)
	if err != nil {
		panic(err)
	}
	return exists
}

func (s *KV[T1, T2]) Delete(key T1) {
	if err := s.TryDelete(key); err != nil {
		panic(err)
	}
}

func (s *KV[T1, T2]) ForEach(fn func(key T1, value T2)) {
	if err := s.TryForEach(fn); err != nil {
		panic(err)
	}
}

func (s *KV[T1, T2]) Clear() {
	if err := s.TryClear(); err != nil {
		panic(err)
	}
}

// StopAllWatchers stops all active watchers
func (s *KV[T1, T2]) StopAllWatchers() {
	if s.watchers == nil {
		return
	}

	s.watchers.mu.Lock()
	defer s.watchers.mu.Unlock()

	for _, w := range s.watchers.watchers {
		w.stop()
	}
	s.watchers.watchers = make(map[string]*watcher[T1, T2])
}

// Helper method to get old value before modification
func (s *KV[T1, T2]) getOldValue(key T1) (T2, bool) {
	var oldValue T2
	var valueStr string
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	sql := fmt.Sprintf("SELECT value FROM %s WHERE key = ?", s.table)
	err := s.db.QueryRowContext(ctx, sql, key).Scan(&valueStr)
	if err != nil {
		return oldValue, false
	}
	
	// Deserialize from JSON
	err = json.Unmarshal([]byte(valueStr), &oldValue)
	if err != nil {
		return oldValue, false
	}

	return oldValue, true
}

// stop safely stops a watcher
func (w *watcher[T1, T2]) stop() {
	w.stopOnce.Do(func() {
		w.stopped = true
		close(w.stopCh)
	})
}

// notify sends events to matching watchers
func (r *watcherRegistry[T1, T2]) notify(key T1, event WatchEvent[T1, T2]) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keyStr := fmt.Sprintf("%v", key)

	for _, w := range r.watchers {
		if w.stopped {
			continue
		}

		// Check if this watcher matches
		matches := false
		if w.key != nil {
			// Exact key match
			matches = fmt.Sprintf("%v", *w.key) == keyStr
		} else if w.prefix != nil {
			// Prefix match
			matches = strings.HasPrefix(keyStr, *w.prefix)
		}

		if matches {
			select {
			case w.ch <- event:
				// Event sent successfully
			case <-w.stopCh:
				// Watcher stopped

				// case <-time.After(100 * time.Millisecond):
				// 	// Don't block if channel is full
			}
		}
	}
}
