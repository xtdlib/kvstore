package kvstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Tx represents a transaction context that provides transactional operations
// All operations within a transaction are atomic - they either all succeed or all fail
type Tx[T1 comparable, T2 comparable] struct {
	tx    *sql.Tx
	table string
	store *KV[T1, T2]
}

// Transaction executes a function within a database transaction
// 
// How it works:
// 1. Starts a database transaction
// 2. Calls your function with a Tx object
// 3. If your function returns an error, all changes are rolled back (cancelled)
// 4. If your function returns nil, all changes are committed (saved)
//
// Example:
//   err := store.Transaction(func(tx *Tx[string, int]) error {
//       // All these operations happen together atomically
//       tx.Set("counter1", 10)
//       tx.Set("counter2", 20)
//       
//       val := tx.Get("counter1")
//       tx.Set("total", val + tx.Get("counter2"))
//       
//       return nil // Success - all changes are saved
//   })
func (s *KV[T1, T2]) Transaction(fn func(tx *Tx[T1, T2]) error) error {
	// Start a database transaction with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	sqlTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	// Create transaction context
	tx := &Tx[T1, T2]{
		tx:    sqlTx,
		table: s.table,
		store: s,
	}
	
	// Execute the user's function
	err = fn(tx)
	
	if err != nil {
		// Something went wrong - rollback all changes
		if rbErr := sqlTx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}
	
	// Everything succeeded - commit all changes
	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	
	return nil
}

// Set stores a key-value pair within the transaction
// Changes are not visible outside the transaction until it commits
func (tx *Tx[T1, T2]) Set(key T1, value T2) error {
	keyBytes, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("failed to marshal key: %w", err)
	}
	
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	
	query := fmt.Sprintf("INSERT OR REPLACE INTO %s (key, value) VALUES (?, ?)", tx.table)
	_, err = tx.tx.Exec(query, string(keyBytes), string(valueBytes))
	return err
}

// Get retrieves a value by key within the transaction
// It sees uncommitted changes made within this transaction
func (tx *Tx[T1, T2]) Get(key T1) (T2, error) {
	var value T2
	var valueStr string
	
	keyBytes, err := json.Marshal(key)
	if err != nil {
		return value, fmt.Errorf("failed to marshal key: %w", err)
	}
	
	query := fmt.Sprintf("SELECT value FROM %s WHERE key = ?", tx.table)
	err = tx.tx.QueryRow(query, string(keyBytes)).Scan(&valueStr)
	if err != nil {
		return value, err
	}
	
	err = json.Unmarshal([]byte(valueStr), &value)
	if err != nil {
		return value, fmt.Errorf("failed to unmarshal value: %w", err)
	}
	
	return value, nil
}

// Delete removes a key-value pair within the transaction
func (tx *Tx[T1, T2]) Delete(key T1) error {
	keyBytes, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("failed to marshal key: %w", err)
	}
	
	query := fmt.Sprintf("DELETE FROM %s WHERE key = ?", tx.table)
	_, err = tx.tx.Exec(query, string(keyBytes))
	return err
}

// Has checks if a key exists within the transaction
func (tx *Tx[T1, T2]) Has(key T1) (bool, error) {
	keyBytes, err := json.Marshal(key)
	if err != nil {
		return false, fmt.Errorf("failed to marshal key: %w", err)
	}
	
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE key = ? LIMIT 1", tx.table)
	var exists int
	err = tx.tx.QueryRow(query, string(keyBytes)).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// ForEach iterates over all key-value pairs within the transaction
func (tx *Tx[T1, T2]) ForEach(fn func(key T1, value T2) error) error {
	query := fmt.Sprintf("SELECT key, value FROM %s ORDER BY key", tx.table)
	rows, err := tx.tx.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	for rows.Next() {
		var keyStr, valueStr string
		if err := rows.Scan(&keyStr, &valueStr); err != nil {
			return err
		}
		
		var key T1
		var value T2
		
		if err := json.Unmarshal([]byte(keyStr), &key); err != nil {
			return fmt.Errorf("failed to unmarshal key: %w", err)
		}
		
		if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
			return fmt.Errorf("failed to unmarshal value: %w", err)
		}
		
		if err := fn(key, value); err != nil {
			return err
		}
	}
	
	return rows.Err()
}

// Clear removes all key-value pairs within the transaction
func (tx *Tx[T1, T2]) Clear() error {
	query := fmt.Sprintf("DELETE FROM %s", tx.table)
	_, err := tx.tx.Exec(query)
	return err
}

// GetOr retrieves a value by key, returning a default if not found
func (tx *Tx[T1, T2]) GetOr(key T1, defaultValue T2) T2 {
	val, err := tx.Get(key)
	if err == sql.ErrNoRows {
		return defaultValue
	}
	if err != nil {
		// In a transaction context, we can't panic, so return default
		return defaultValue
	}
	return val
}