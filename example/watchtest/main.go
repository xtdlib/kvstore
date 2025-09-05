package main

import (
	"fmt"

	"github.com/xtdlib/kvstore"
)

func main() {
	// Create store
	store, err := kvstore.NewAt[string, string]("./xxx.db", "watchdemo")
	if err != nil {
		panic(err)
	}

	// Start watching key "message" - returns channel and cancel function
	events, cancel := store.Watch("message")

	go func() {
		// Make some changes
		store.Set("message", "hello")
		store.Set("message", "world")
		store.Delete("message")
		cancel() // This will close the channel
	}()

	for event := range events {
		fmt.Printf("Change detected: key=%s value=%s\n", event.Key, event.Value)
	}
}
