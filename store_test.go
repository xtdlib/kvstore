package kv2_test

import (
	"os"
	"testing"

	"github.com/xtdlib/kv2"
)

func TestNew(t *testing.T) {
	dbFile := "test.db"
	defer os.Remove(dbFile)

	kv := kv2.New[string, string](dbFile)
	if kv == nil {
		t.Fatal("New returned nil")
	}
}

func TestSetGet(t *testing.T) {
	dbFile := "test.db"
	defer os.Remove(dbFile)

	kv := kv2.New[string, string](dbFile)

	err := kv.Set("key1", "value1")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := kv.Get("key1", "")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value1" {
		t.Fatalf("Got %s, expected value1", val)
	}
}

func TestSetReplace(t *testing.T) {
	dbFile := "test.db"
	defer os.Remove(dbFile)

	kv := kv2.New[string, string](dbFile)

	err := kv.Set("key1", "value1")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	err = kv.Set("key1", "value2")
	if err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	val, err := kv.Get("key1", "")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value2" {
		t.Fatalf("Got %s, expected value2", val)
	}
}

func TestGetNotFound(t *testing.T) {
	dbFile := "test.db"
	defer os.Remove(dbFile)

	kv := kv2.New[string, string](dbFile)

	_, err := kv.Get("nonexistent", "")
	if err == nil {
		t.Fatal("Expected error for nonexistent key")
	}
}

func TestIntKeys(t *testing.T) {
	dbFile := "test.db"
	defer os.Remove(dbFile)

	kv := kv2.New[int, string](dbFile)

	err := kv.Set(42, "answer")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := kv.Get(42, "")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "answer" {
		t.Fatalf("Got %s, expected answer", val)
	}
}

func TestDelete(t *testing.T) {
	dbFile := "test.db"
	defer os.Remove(dbFile)

	kv := kv2.New[string, string](dbFile)

	err := kv.Set("key1", "value1")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	err = kv.Delete("key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = kv.Get("key1", "")
	if err == nil {
		t.Fatal("Expected error after delete")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	dbFile := "test.db"
	defer os.Remove(dbFile)

	kv := kv2.New[string, string](dbFile)

	err := kv.Delete("nonexistent")
	if err != nil {
		t.Fatalf("Delete nonexistent key failed: %v", err)
	}
}