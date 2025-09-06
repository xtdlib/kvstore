package main

import (
	"log"
	"time"

	"github.com/xtdlib/kvstore"
)

type X struct {
	Name string
	Age  int
}

func main() {
	store := kvstore.New[string, time.Time]("time")
	store.Clear()
	store.Set("last", time.Now())
	log.Println(store.Get("last"))
}
