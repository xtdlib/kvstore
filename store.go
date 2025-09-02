package kv2

import (
	"database/sql"

	_ "modernc.org/sqlite"
)


type KV[T1 any, T2 any] struct {
	db *sql.DB
}

func New[T1 any, T2 any](name string) *KV[T1, T2] {
	store := &KV[T1, T2]{}
	var err error
	store.db, err = sql.Open("sqlite", name)
	if err != nil {
		panic(err)
	}

	store.db.Exec("CREATE TABLE IF NOT EXISTS store (key TEXT PRIMARY KEY, value TEXT)")
	return store
}

func (s *KV[T1, T2]) Set(key T1, value T2) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO store (key, value) VALUES (?, ?)", key, value)
	if err != nil {
		return err
	}
	return nil
}

func (s *KV[T1, T2]) Get(key T1, value T2) (T2, error) {
	var v T2
	err := s.db.QueryRow("SELECT value FROM store WHERE key = ?", key).Scan(&v)
	return v, err
}

func (s *KV[T1, T2]) Delete(key T1) error {
	_, err := s.db.Exec("DELETE FROM store WHERE key = ?", key)
	return err
}
