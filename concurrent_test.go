package kvstore_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/xtdlib/kvstore"
)

func TestConcurrentReads(t *testing.T) {
	kv := kvstore.New[string, string]("test_concurrent_reads")

	// Populate with test data
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		if err := kv.Set(key, value); err != nil {
			t.Fatalf("Failed to set %s: %v", key, err)
		}
	}

	// Concurrent reads
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key%d", j)
				expected := fmt.Sprintf("value%d", j)
				val, err := kv.Get(key, "")
				if err != nil {
					errors <- fmt.Errorf("worker %d: get %s failed: %v", worker, key, err)
					return
				}
				if val != expected {
					errors <- fmt.Errorf("worker %d: got %s, expected %s", worker, val, expected)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Fatal(err)
	}
}

func TestConcurrentWrites(t *testing.T) {
	kv := kvstore.New[int, int]("test_concurrent_writes")

	// Concurrent writes from multiple goroutines
	var wg sync.WaitGroup
	errors := make(chan error, 100)
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := worker*100 + j
				value := key * 2
				if err := kv.Set(key, value); err != nil {
					errors <- fmt.Errorf("worker %d: set %d failed: %v", worker, key, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Fatal(err)
	}

	// Verify all writes succeeded
	for i := 0; i < 1000; i++ {
		expected := i * 2
		val, err := kv.Get(i, 0)
		if err != nil {
			t.Fatalf("Get %d failed: %v", i, err)
		}
		if val != expected {
			t.Fatalf("For key %d: got %d, expected %d", i, val, expected)
		}
	}
}

func TestConcurrentMixed(t *testing.T) {
	kv := kvstore.New[string, int]("test_concurrent_mixed")

	// Initialize some data
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("key%d", i)
		if err := kv.Set(key, i); err != nil {
			t.Fatalf("Failed to set %s: %v", key, err)
		}
	}

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				key := fmt.Sprintf("new_key_%d_%d", worker, j)
				if err := kv.Set(key, worker*100+j); err != nil {
					errors <- fmt.Errorf("writer %d: set %s failed: %v", worker, key, err)
				}
			}
		}(i)
	}

	// Readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				key := fmt.Sprintf("key%d", j)
				_, err := kv.Get(key, 0)
				if err != nil {
					errors <- fmt.Errorf("reader %d: get %s failed: %v", worker, key, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Fatal(err)
	}
}

func TestMultipleTablesConcurrent(t *testing.T) {
	kv1 := kvstore.New[string, string]("table1")
	kv2 := kvstore.New[string, string]("table2")

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Write to table1
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("t1_key%d", i)
			value := fmt.Sprintf("t1_value%d", i)
			if err := kv1.Set(key, value); err != nil {
				errors <- fmt.Errorf("table1 set failed: %v", err)
			}
		}
	}()

	// Write to table2
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("t2_key%d", i)
			value := fmt.Sprintf("t2_value%d", i)
			if err := kv2.Set(key, value); err != nil {
				errors <- fmt.Errorf("table2 set failed: %v", err)
			}
		}
	}()

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Fatal(err)
	}

	// Verify data isolation between tables
	_, err1 := kv1.Get("t2_key0", "")
	if err1 == nil {
		t.Fatal("Table isolation failed: found table2 key in table1")
	}

	_, err2 := kv2.Get("t1_key0", "")
	if err2 == nil {
		t.Fatal("Table isolation failed: found table1 key in table2")
	}
}