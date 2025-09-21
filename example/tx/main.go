package main

import (
	"log"

	"errors"

	"github.com/xtdlib/kvstore"
)

func main() {
	store := kvstore.New[string, int]("mystore")
	store.Clear()

	// Simple transaction - all these happen together
	store.Transaction(func(tx *kvstore.Tx[string, int]) error {
		tx.Set("apples", 5)
		tx.Set("oranges", 3)
		tx.Set("total", 8)
		return errors.New("oops, something went wrong")
	})

	log.Println(store.Get("apples"))
}
