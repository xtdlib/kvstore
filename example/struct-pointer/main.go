package main

import (
	"log"

	"github.com/xtdlib/kvstore"
)

type X struct {
	Name string
	Age  int
}

func main() {
	// store := kvstore.New[string, sql.NullTime]("time")
	store := kvstore.New[string, *X]("time")
	store.Clear()
	store.Set("last", &X{Name: "john", Age: 18})
	log.Println(store.Get("last"))

	// now := time.Now()
	// _ = now
	// log.Println("balance:", balance)
}
