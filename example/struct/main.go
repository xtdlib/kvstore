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
	store := kvstore.New[string, X]("struct_example")
	store.Clear()
	
	// Test setting multiple structs
	store.Set("user1", X{Name: "john", Age: 18})
	store.Set("user2", X{Name: "alice", Age: 25})
	store.Set("user3", X{Name: "bob", Age: 30})
	
	// Test getting values
	log.Println("user1:", store.Get("user1"))
	log.Println("user2:", store.Get("user2"))
	
	// Test ForEach
	log.Println("\nAll users:")
	store.ForEach(func(key string, value X) {
		log.Printf("  %s: %+v", key, value)
	})
	
	// Test Has
	log.Println("\nHas user1:", store.Has("user1"))
	log.Println("Has user4:", store.Has("user4"))
	
	// Test Delete
	store.Delete("user2")
	log.Println("\nAfter deleting user2:")
	store.ForEach(func(key string, value X) {
		log.Printf("  %s: %+v", key, value)
	})
}
