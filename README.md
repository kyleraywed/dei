# Dei

An "easy enough" **d**eferred-**e**xecution **i**terator library.

```go
func (iter *Dei[T]) Filter(in func(value T) bool, comments ...string)
func (iter *Dei[T]) Map(in func(value T) T, comments ...string)
func (iter *Dei[T]) Take(in int)
func (iter *Dei[T]) Skip(in int)

func (iter *Dei[T]) Apply(in []T) []T
```

Usage

```go
package main

import (
    "github.com/kyleraywed/dei"
)

func main() {
    var iter dei.Dei[int]

    iter.Filter(func(value int) bool {
        return value % 2 == 0
    }, "Get just the evens")

    iter.Map(func(value int) {
        return value * 2
    }, "Double them")

    iter.Filter(func(value int) bool { // get just values > 10
        return value > 10
    })

    iter.Take(2) // get just the first 2 elements

    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    new := iter.Apply(numbers) // []int{16, 18}
}
```

- Notes:
    -
    - The order of operations is the order they are registered. 