package main

import (
	"log"

	"github.com/xtdlib/kvstore"
)

func main() {
	store := kvstore.New[string, []string]("time")
	store.Clear()
	store.Set("last", []string{"one", "two", "three"})
	log.Println(store.Get("last"))

	// now := time.Now()
	// _ = now
	// log.Println("balance:", balance)
}
