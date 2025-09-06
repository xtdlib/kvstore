package main

import (
	"log"

	"github.com/xtdlib/kvstore"
	"github.com/xtdlib/rat"
)

func main() {
	store := kvstore.New[*rat.Rational, *rat.Rational]("rat")

	store.Clear()
	store.Set(rat.Rat(1), rat.Rat(3))
	store.Set(rat.Rat("1/3"), rat.Rat(5))
	if !store.Get(rat.Rat(1)).Equal(3) {
		panic("should be same")
	}

	ksum := rat.Rat(0)
	store.ForEach(func(k *rat.Rational, v *rat.Rational) {
		ksum = ksum.Add(k)
	})
	ksum.SetPrecision(9)

	if !ksum.Equal("4/3") {
		panic("should be same")
	} else {
		log.Println("ksum:", ksum)
		log.Println("ksum:", ksum.DecimalString())
	}

}
