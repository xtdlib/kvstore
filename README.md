# kvstore

Simple kv store backed by sqlite. By default it creates db file in `$XDG_CACHE_HOME/$0/$0.db`

```
package main

import (
	"github.com/xtdlib/kvstore"
	"github.com/xtdlib/rat"
)

func main() {
	store := kvstore.New[string, *rat.Rational]("balance")

	store.Set("account1", rat.Rat(0.2))
	store.Set("account2", rat.Rat(0.1))

	sum := rat.Rat(0)
	store.ForEach(func(key string, value *rat.Rational) {
		sum = sum.Add(value)
	})
	println("sum: " + sum.String()) // sum: 0.3
}
```

## API

- `New[K, V]`
- `Set(K, V)`
- `SetIf(K, V)`
- `Get(K)`
- `GetOr(K, V)`
- `Has(K)` bool
- `Delete(K)`
- `Clear()`
- `ForEach(func(K, V))`

