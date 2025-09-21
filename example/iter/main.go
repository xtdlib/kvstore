package main

import (
	"fmt"

	"github.com/xtdlib/kvstore"
)

func main() {
	// Create a new store
	store := kvstore.New[string, int]("example")

	// Add some data
	store.Set("apple", 5)
	store.Set("banana", 3)
	store.Set("cherry", 8)

	for k, v := range store.All {
		fmt.Printf("%s: %d\n", k, v)
	}

	for k, v := range store.Backward {
		fmt.Printf("%s: %d\n", k, v)
	}
}
