# DEI

A bare-bones **d**eferred-**e**xecution **i**terator library.

```go
// Keeps only the elements where in returns true. Optional comment strings.
func (iter *Dei[T]) Filter(in func(value T) bool, comments ...string)

// Transform each element by applying a function. Optional comment strings.
func (iter *Dei[T]) Map(in func(value T) T, comments ...string)

// Yield only the first n items from the iterator. Comments inferred.
func (iter *Dei[T]) Take(n int)

// Skip the first n items and yields the rest. Comments inferred.
func (iter *Dei[T]) Skip(n int)

// Interpret orders on data. Return new slice.
func (iter *Dei[T]) Apply(input []T) []T
```

Usage

```go
package main

import (
    "github.com/kyleraywed/dei"
)

func main() {
    // create a Dei object of the type you want to iterate over
    var iter dei.Dei[int]

    // Each adapter is applied in the order in which they are declared.
    // So this would happen first.
    iter.Filter(func(value int) bool {
        return value % 2 == 0
    }, "Get just the evens")

    // Second. Also notice the optional comment. These can be viewed via String()
    iter.Map(func(value int) int {
        return value * 2
    }, "Double them")

    // Third.
    iter.Filter(func(value int) bool { // get just values > 10
        return value > 10
    })

    // Last. Take and skip still log inferred comments.
    iter.Take(2) // get just the first 2 elements

    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    new := iter.Apply(numbers) // []int{12, 16}
}
```