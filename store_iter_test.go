package kvstore_test

import (
	"os"
	"testing"

	"github.com/xtdlib/kvstore"
)

func TestIter(t *testing.T) {
	// Create temp database
	dbPath := "test_iter.db"
	defer os.Remove(dbPath)

	store := kvstore.New[string, int]("test_iter")

	// Add test data
	testData := map[string]int{
		"apple":  1,
		"banana": 2,
		"cherry": 3,
		"date":   4,
		"elder":  5,
	}

	for k, v := range testData {
		store.Set(k, v)
	}

	// Test forward iterator
	t.Run("Forward Iterator", func(t *testing.T) {
		collected := make(map[string]int)
		count := 0
		var lastKey string

		for k, v := range store.Iter() {
			collected[k] = v
			count++
			if lastKey != "" && k < lastKey {
				t.Errorf("Keys not in ascending order: %s came after %s", k, lastKey)
			}
			lastKey = k
		}

		if count != len(testData) {
			t.Errorf("Expected %d items, got %d", len(testData), count)
		}

		for k, expectedV := range testData {
			if v, ok := collected[k]; !ok {
				t.Errorf("Missing key: %s", k)
			} else if v != expectedV {
				t.Errorf("Wrong value for %s: got %d, want %d", k, v, expectedV)
			}
		}
	})

	// Test reverse iterator
	t.Run("Reverse Iterator", func(t *testing.T) {
		collected := make(map[string]int)
		count := 0
		var lastKey string

		for k, v := range store.IterReverse() {
			collected[k] = v
			count++
			if lastKey != "" && k > lastKey {
				t.Errorf("Keys not in descending order: %s came after %s", k, lastKey)
			}
			lastKey = k
		}

		if count != len(testData) {
			t.Errorf("Expected %d items, got %d", len(testData), count)
		}

		for k, expectedV := range testData {
			if v, ok := collected[k]; !ok {
				t.Errorf("Missing key: %s", k)
			} else if v != expectedV {
				t.Errorf("Wrong value for %s: got %d, want %d", k, v, expectedV)
			}
		}
	})

	// Test early break
	t.Run("Early Break", func(t *testing.T) {
		count := 0
		for range store.Iter() {
			count++
			if count == 3 {
				break
			}
		}

		if count != 3 {
			t.Errorf("Expected to break at 3, got %d", count)
		}
	})

	// Test empty store
	t.Run("Empty Store", func(t *testing.T) {
		emptyStore := kvstore.New[string, int]("empty_iter")

		count := 0
		for range emptyStore.Iter() {
			count++
		}

		if count != 0 {
			t.Errorf("Expected 0 items in empty store, got %d", count)
		}
	})
}
