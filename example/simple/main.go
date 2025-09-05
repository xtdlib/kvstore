package main

import (
	"log"

	"github.com/xtdlib/kvstore"
	"github.com/xtdlib/rat"
)

func main() {
	store := kvstore.New[string, *rat.Rational]("balance")

	balance := store.GetOr("accountx", rat.Rat(3))
	log.Println("balance:", balance)
}
