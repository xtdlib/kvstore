package main

import (
	"github.com/xtdlib/kvstore"
	"github.com/xtdlib/rat"
)

func main() {
	store := kvstore.New[string, *rat.Rational]("balance")
	store.Clear()

	store.Set("account1", rat.Rat("0.2"))
	store.Set("account2", rat.Rat("0.1"))

	sum := rat.Rat(0)

	println("sum: " + sum.String()) // sum: 0.3
	if !sum.Equal(rat.Rat("0.3")) {
		panic("sum is not 0.3")
	}
}
