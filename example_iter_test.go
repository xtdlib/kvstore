package kvstore_test

import (
	"fmt"

	"github.com/xtdlib/kvstore"
)

func ExampleKV_Iter() {
	// Create a new store
	store := kvstore.New[string, int]("example")

	// Add some data
	store.Set("apple", 5)
	store.Set("banana", 3)
	store.Set("cherry", 8)

	// Iterate using range-over-func (Go 1.23+)
	fmt.Println("Forward iteration:")
	for key, value := range store.Iter() {
		fmt.Printf("%s: %d\n", key, value)
	}

	// Output:
	// Forward iteration:
	// apple: 5
	// banana: 3
	// cherry: 8
}

func ExampleKV_IterReverse() {
	// Create a new store
	store := kvstore.New[string, int]("example_reverse")

	// Add some data
	store.Set("apple", 5)
	store.Set("banana", 3)
	store.Set("cherry", 8)

	// Iterate in reverse order
	fmt.Println("Reverse iteration:")
	for key, value := range store.IterReverse() {
		fmt.Printf("%s: %d\n", key, value)
	}

	// Output:
	// Reverse iteration:
	// cherry: 8
	// banana: 3
	// apple: 5
}

