package main

import (
	"fmt"

	"github.com/xtdlib/kvstore"
)

func main() {
	store := kvstore.New[string, string]("example")

	store.Set("name", "John")
	store.Set("city", "New York")

	val, err := store.Get("name", "")
	if err != nil {
		panic(err)
	}
	fmt.Println("name:", val)

	store.Delete("city")
}
