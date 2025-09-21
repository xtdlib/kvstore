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

// Test panic versions (without Try prefix)
func TestSetGet(t *testing.T) {
	kv := kvstore.New[string, string]("test_setget")

	kv.Set("key1", "value1")

	val := kv.Get("key1")
	if val != "value1" {
		t.Fatalf("Got %s, expected value1", val)
	}
}

func TestSetReplace(t *testing.T) {
	kv := kvstore.New[string, string]("test_replace")

	kv.Set("key1", "value1")
	kv.Set("key1", "value2")

	val := kv.Get("key1")
	if val != "value2" {
		t.Fatalf("Got %s, expected value2", val)
	}
}

func TestGetNotFound(t *testing.T) {
	kv := kvstore.New[string, string]("test_notfound")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for nonexistent key")
		}
	}()

	kv.Get("nonexistent")
}

func TestIntKeys(t *testing.T) {
	kv := kvstore.New[int, string]("test_intkeys")

	kv.Set(42, "answer")

	val := kv.Get(42)
	if val != "answer" {
		t.Fatalf("Got %s, expected answer", val)
	}
}

func TestHas(t *testing.T) {
	kv := kvstore.New[string, string]("test_has")

	// Test non-existent key
	exists := kv.Has("key1")
	if exists {
		t.Fatal("Expected false for non-existent key")
	}

	// Add a key
	kv.Set("key1", "value1")

	// Test existing key
	exists = kv.Has("key1")
	if !exists {
		t.Fatal("Expected true for existing key")
	}

	// Delete the key
	kv.Delete("key1")

	// Test after deletion
	exists = kv.Has("key1")
	if exists {
		t.Fatal("Expected false after deletion")
	}
}

func TestDelete(t *testing.T) {
	kv := kvstore.New[string, string]("test_delete")

	kv.Set("key1", "value1")
	kv.Delete("key1")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic after delete")
		}
	}()

	kv.Get("key1")
}

func TestDeleteNonExistent(t *testing.T) {
	kv := kvstore.New[string, string]("test_delete_nonexistent")

	// Should not panic
	kv.Delete("nonexistent")
}

func TestClear(t *testing.T) {
	kv := kvstore.New[string, string]("test_clear")

	// Add test data
	kv.Set("key1", "value1")
	kv.Set("key2", "value2")
	kv.Set("key3", "value3")

	// Verify data exists
	val := kv.Get("key1")
	if val != "value1" {
		t.Fatalf("Got %s, expected value1", val)
	}

	// Clear all data
	kv.Clear()

	// Verify all data is gone
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for key1 after clear")
		}
	}()
	kv.Get("key1")
}

func TestClearEmpty(t *testing.T) {
	kv := kvstore.New[string, string]("test_clear_empty")

	// Clear empty store should not panic
	kv.Clear()
}

// Test Try versions (with error handling)
func TestTrySetGet(t *testing.T) {
	kv := kvstore.New[string, string]("test_try_setget")

	_, err := kv.TrySet("key1", "value1")
	if err != nil {
		t.Fatalf("TrySet failed: %v", err)
	}

	val, err := kv.TryGet("key1")
	if err != nil {
		t.Fatalf("TryGet failed: %v", err)
	}
	if val != "value1" {
		t.Fatalf("Got %s, expected value1", val)
	}
}

func TestTryGetNotFound(t *testing.T) {
	kv := kvstore.New[string, string]("test_try_notfound")

	_, err := kv.TryGet("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent key")
	}
}

func TestTryHas(t *testing.T) {
	kv := kvstore.New[string, string]("test_try_has")

	// Test non-existent key
	exists, err := kv.TryHas("key1")
	if err != nil {
		t.Fatalf("TryHas failed: %v", err)
	}
	if exists {
		t.Fatal("Expected false for non-existent key")
	}

	// Add a key
	_, err = kv.TrySet("key1", "value1")
	if err != nil {
		t.Fatalf("TrySet failed: %v", err)
	}

	// Test existing key
	exists, err = kv.TryHas("key1")
	if err != nil {
		t.Fatalf("TryHas failed: %v", err)
	}
	if !exists {
		t.Fatal("Expected true for existing key")
	}
}

func TestTryDelete(t *testing.T) {
	kv := kvstore.New[string, string]("test_try_delete")

	_, err := kv.TrySet("key1", "value1")
	if err != nil {
		t.Fatalf("TrySet failed: %v", err)
	}

	err = kv.TryDelete("key1")
	if err != nil {
		t.Fatalf("TryDelete failed: %v", err)
	}

	_, err = kv.TryGet("key1")
	if err == nil {
		t.Fatal("Expected error after delete")
	}
}

func TestTryClear(t *testing.T) {
	kv := kvstore.New[string, string]("test_try_clear")

	// Add test data
	kv.TrySet("key1", "value1")
	kv.TrySet("key2", "value2")

	// Clear all data
	err := kv.TryClear()
	if err != nil {
		t.Fatalf("TryClear failed: %v", err)
	}

	// Verify all data is gone
	_, err = kv.TryGet("key1")
	if err == nil {
		t.Fatal("Expected error for key1 after clear")
	}
}

func TestGetOr(t *testing.T) {
	kv := kvstore.New[string, string]("test_getor")

	// Test with non-existent key - should return default value
	defaultVal := "default_value"
	val := kv.GetOr("nonexistent", defaultVal)
	if val != defaultVal {
		t.Fatalf("Got %s, expected %s", val, defaultVal)
	}

	// Add a key
	kv.Set("key1", "actual_value")

	// Test with existing key - should return actual value
	val = kv.GetOr("key1", defaultVal)
	if val != "actual_value" {
		t.Fatalf("Got %s, expected actual_value", val)
	}

	// Test with different types (int)
	kvInt := kvstore.New[string, int]("test_getor_int")

	// Non-existent key returns default
	intVal := kvInt.GetOr("missing", 42)
	if intVal != 42 {
		t.Fatalf("Got %d, expected 42", intVal)
	}

	// Existing key returns actual value
	kvInt.Set("answer", 100)
	intVal = kvInt.GetOr("answer", 42)
	if intVal != 100 {
		t.Fatalf("Got %d, expected 100", intVal)
	}
}

