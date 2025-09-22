package main

import (
	"fmt"

	"github.com/xtdlib/kvstore"
)

func main() {
	// Create a new store
	store := kvstore.New[string, int]("example")
	store.Clear() // Start fresh

	// Add some data
	store.Set("zebra", 5)
	store.Set("apple", 3)
	store.Set("banana", 8)
	store.Set("cherry", 2)

	fmt.Println("=== Keys Forward Iteration ===")
	store.Keys(func(key string) bool {
		fmt.Printf("Key: %s, Value: %d\n", key, store.Get(key))
		return true
	})

	fmt.Println("\n=== Keys Backward Iteration ===")
	store.KeysBackward(func(key string) bool {
		fmt.Printf("Key: %s, Value: %d\n", key, store.Get(key))
		return true
	})

	fmt.Println("\n=== Keys with Early Termination ===")
	count := 0
	store.Keys(func(key string) bool {
		fmt.Printf("Key: %s\n", key)
		count++
		return count < 2 // Stop after 2 keys
	})

	fmt.Println("\n=== Go 1.23+ Range-over-func Keys Forward ===")
	for key := range store.KeysIter() {
		fmt.Printf("Key: %s, Value: %d\n", key, store.Get(key))
	}

	fmt.Println("\n=== Go 1.23+ Range-over-func Keys Backward ===")
	for key := range store.KeysIterReverse() {
		fmt.Printf("Key: %s, Value: %d\n", key, store.Get(key))
	}

	fmt.Println("\n=== Compare with Full Iteration ===")
	fmt.Println("All key-value pairs (forward):")
	store.All(func(key string, value int) bool {
		fmt.Printf("  %s: %d\n", key, value)
		return true
	})

	fmt.Println("All key-value pairs (backward):")
	store.Backward(func(key string, value int) bool {
		fmt.Printf("  %s: %d\n", key, value)
		return true
	})

	fmt.Println("All key-value pairs (backward):")
	for k, v := range store.Backward {
		fmt.Printf("  %s: %d\n", k, v)
	}
}
