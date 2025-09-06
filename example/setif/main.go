package main

import (
	"errors"
	"log"

	"github.com/xtdlib/kvstore"
	"github.com/xtdlib/rat"
	"github.com/xtdlib/try"
)

func FailFunc() (*rat.Rational, error) {
	return nil, errors.New("fail")
}

func main() {
	store := kvstore.New[string, *rat.Rational]("rat")

	store.Clear()
	store.SetIf("fail", try.L1(FailFunc()))

	log.Println(store.Has("fail"))
}
