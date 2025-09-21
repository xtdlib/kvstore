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

	// go func() {
	// 	time.Sleep(time.Second)
	// 	store.Set("date", 10)
	// 	log.Println("date", store.Get("date"))
	// }()
	//
	// // Iterate using range-over-func (Go 1.23+)
	// fmt.Println("Forward iteration:")
	// for key, value := range store.Iter() {
	// 	time.Sleep(time.Second)
	// }

	for k, v := range store.All {
		fmt.Printf("%s: %d\n", k, v)
	}

	for k, v := range store.Backward {
		fmt.Printf("%s: %d\n", k, v)
	}

	it := store.Backward
	// for it.Next() {
	// }

	// slices.Collect(store.Iter)

	// Output:
	// Forward iteration:
	// apple: 5
	// banana: 3
	// cherry: 8
}
