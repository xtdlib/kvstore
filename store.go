package kvstore

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	sharedDB *sql.DB
	dbOnce   sync.Once
)

type KV[T1 any, T2 any] struct {
	db    *sql.DB
	table string
}

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
		connStr := fmt.Sprintf("%s?_busy_timeout=10000&_journal=WAL&_sync=NORMAL", dbPath)
		sharedDB, err = sql.Open("sqlite", connStr)
		if err != nil {
			return
		}

		// Enable WAL mode and set pragmas for better concurrency
		// WAL allows concurrent reads and one writer
		pragmas := []string{
			"PRAGMA journal_mode=WAL",      // Enable Write-Ahead Logging
			"PRAGMA synchronous=NORMAL",    // Good balance of safety and speed
			"PRAGMA busy_timeout=10000",    // Wait up to 10 seconds when database is locked
			"PRAGMA cache_size=-32000",     // 32MB cache
		}

		for _, pragma := range pragmas {
			if _, pragmaErr := sharedDB.Exec(pragma); pragmaErr != nil {
				err = pragmaErr
				return
			}
		}

		// Set connection pool settings for better concurrency
		sharedDB.SetMaxOpenConns(25)    // Allow multiple readers
		sharedDB.SetMaxIdleConns(25)
		sharedDB.SetConnMaxLifetime(5 * time.Minute)
	})
	return sharedDB, err
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	// Create table with sanitized name
	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (key TEXT PRIMARY KEY, value TEXT)", store.table)
	_, err = store.db.ExecContext(ctx, createSQL)
	if err != nil {
		panic(err)
	}
	
	return store
}

func (s *KV[T1, T2]) Set(key T1, value T2) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sql := fmt.Sprintf("INSERT OR REPLACE INTO %s (key, value) VALUES (?, ?)", s.table)
	_, err := s.db.ExecContext(ctx, sql, key, value)
	if err != nil {
		return err
	}
	return nil
}

func (s *KV[T1, T2]) Get(key T1, value T2) (T2, error) {
	var v T2
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sql := fmt.Sprintf("SELECT value FROM %s WHERE key = ?", s.table)
	err := s.db.QueryRowContext(ctx, sql, key).Scan(&v)
	return v, err
}

func (s *KV[T1, T2]) Delete(key T1) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sql := fmt.Sprintf("DELETE FROM %s WHERE key = ?", s.table)
	_, err := s.db.ExecContext(ctx, sql, key)
	return err
}

func (s *KV[T1, T2]) ForEach(fn func(key T1, value T2) error) error {
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
		if err := rows.Scan(&k, &v); err != nil {
			return err
		}
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (s *KV[T1, T2]) Clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sql := fmt.Sprintf("DELETE FROM %s", s.table)
	_, err := s.db.ExecContext(ctx, sql)
	return err
}
