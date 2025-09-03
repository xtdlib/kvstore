# kvstore

```
package main

import (
	"github.com/xtdlib/kvstore"
	"github.com/xtdlib/rat"
)

func main() {
	store := kvstore.New[string, *rat.Rational]("example")

	store.Set("account1", rat.Rat(0.2))
	store.Set("account2", rat.Rat(0.1))

	sum := rat.Rat(0)
	store.ForEach(func(key string, value *rat.Rational) error {
		sum = sum.Add(value)
		return nil
	})
	println("sum: " + sum.String()) // sum: 0.3
}
```

