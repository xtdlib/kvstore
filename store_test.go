package kvstore_test

import (
	"testing"

	"github.com/xtdlib/kvstore"
)

func TestNew(t *testing.T) {
	kv := kvstore.New[string, string]("test_new")
	if kv == nil {
		t.Fatal("New returned nil")
	}
}

func TestSetGet(t *testing.T) {
	kv := kvstore.New[string, string]("test_setget")

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
	kv := kvstore.New[string, string]("test_replace")

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
	kv := kvstore.New[string, string]("test_notfound")

	_, err := kv.Get("nonexistent", "")
	if err == nil {
		t.Fatal("Expected error for nonexistent key")
	}
}

func TestIntKeys(t *testing.T) {
	kv := kvstore.New[int, string]("test_intkeys")

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
	kv := kvstore.New[string, string]("test_delete")

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
	kv := kvstore.New[string, string]("test_delete_nonexistent")

	err := kv.Delete("nonexistent")
	if err != nil {
		t.Fatalf("Delete nonexistent key failed: %v", err)
	}
}

func TestForEach(t *testing.T) {
	kv := kvstore.New[string, int]("test_foreach")

	// Add test data
	kv.Set("a", 1)
	kv.Set("b", 2)
	kv.Set("c", 3)

	// Count items
	count := 0
	sum := 0
	err := kv.ForEach(func(key string, value int) error {
		count++
		sum += value
		return nil
	})

	if err != nil {
		t.Fatalf("ForEach failed: %v", err)
	}
	if count != 3 {
		t.Fatalf("Expected 3 items, got %d", count)
	}
	if sum != 6 {
		t.Fatalf("Expected sum of 6, got %d", sum)
	}
}

func TestClear(t *testing.T) {
	kv := kvstore.New[string, string]("test_clear")

	// Add test data
	kv.Set("key1", "value1")
	kv.Set("key2", "value2")
	kv.Set("key3", "value3")

	// Verify data exists
	val, err := kv.Get("key1", "")
	if err != nil {
		t.Fatalf("Get before clear failed: %v", err)
	}
	if val != "value1" {
		t.Fatalf("Got %s, expected value1", val)
	}

	// Clear all data
	err = kv.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify all data is gone
	_, err = kv.Get("key1", "")
	if err == nil {
		t.Fatal("Expected error for key1 after clear")
	}
	_, err = kv.Get("key2", "")
	if err == nil {
		t.Fatal("Expected error for key2 after clear")
	}
	_, err = kv.Get("key3", "")
	if err == nil {
		t.Fatal("Expected error for key3 after clear")
	}

	// Verify we can add new data after clear
	err = kv.Set("newkey", "newvalue")
	if err != nil {
		t.Fatalf("Set after clear failed: %v", err)
	}
	val, err = kv.Get("newkey", "")
	if err != nil {
		t.Fatalf("Get after clear failed: %v", err)
	}
	if val != "newvalue" {
		t.Fatalf("Got %s, expected newvalue", val)
	}
}

func TestClearEmpty(t *testing.T) {
	kv := kvstore.New[string, string]("test_clear_empty")

	// Clear empty store should not error
	err := kv.Clear()
	if err != nil {
		t.Fatalf("Clear empty store failed: %v", err)
	}
}
