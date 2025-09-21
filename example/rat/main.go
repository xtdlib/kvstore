package main

import (
	"encoding/json"
	"log"

	"github.com/xtdlib/kvstore"
	"github.com/xtdlib/rat"
	"github.com/xtdlib/try"
)

func main() {
	store := kvstore.New[*rat.Rational, *rat.Rational]("rat")

	store.Clear()
	store.Set(rat.Rat(1), rat.Rat(3))
	if !store.Get(rat.Rat(1)).Equal(3) {
		panic("should be 3")
	}

	log.Println("json", string(try.E1(json.Marshal(rat.Rat("1/3")))))

	store.Set(rat.Rat("1/3"), rat.Rat(5))
	ksum := rat.Rat(0)

	if !ksum.Equal("4/3") {
		log.Println("ksum:", ksum.FractionString())
		panic("should be same")
	} else {
		log.Println("ksum:", ksum)
		log.Println("ksum:", ksum.DecimalString())
	}

}
