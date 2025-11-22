package dei

import (
	"fmt"
	"log"
	"strconv"
)

type order struct {
	method   string
	index    int
	comments []string
}

type Dei[T any] struct {
	filters    []func(t T) bool
	mappers    []func(t T) T
	takeCounts []int
	skipCounts []int

	orders []order
}

// Keep only the elements where in returns true. Optional comment strings.
func (iter *Dei[T]) Filter(in func(value T) bool, comments ...string) {
	iter.filters = append(iter.filters, in)
	iter.orders = append(iter.orders, order{
		method: "filter", index: len(iter.filters) - 1, comments: comments,
	})
}

// Transform each element by applying a function. Optional comment strings.
func (iter *Dei[T]) Map(in func(value T) T, comments ...string) {
	iter.mappers = append(iter.mappers, in)
	iter.orders = append(iter.orders, order{
		method: "map", index: len(iter.mappers) - 1, comments: comments,
	})
}

// Yield only the first n items from the iterator. Comments inferred.
func (iter *Dei[T]) Take(n int) {
	if n < 1 {
		log.Printf("Take(%v): No order submitted.", n)
		return
	}

	iter.takeCounts = append(iter.takeCounts, n)

	iter.orders = append(iter.orders, order{
		method: "take", index: len(iter.takeCounts) - 1, comments: []string{strconv.Itoa(n)},
	})
}

// Skip the first n items and yields the rest. Comments inferred.
func (iter *Dei[T]) Skip(n int) {
	if n < 1 {
		log.Printf("Skip(%v): No order submitted.", n)
		return
	}

	iter.skipCounts = append(iter.skipCounts, n)

	iter.orders = append(iter.orders, order{
		method: "skip", index: len(iter.skipCounts) - 1, comments: []string{strconv.Itoa(n)},
	})
}

// Interpret orders on data. Return new slice.
func (iter *Dei[T]) Apply(input []T) []T {
	workingSlice := input

	for _, order := range iter.orders {
		switch order.method {

		case "filter":
			tempSlice := make([]T, 0, len(workingSlice))
			instruct := iter.filters[order.index]

			for _, val := range workingSlice {
				if instruct(val) {
					tempSlice = append(tempSlice, val)
				}
			}

			workingSlice = tempSlice

		case "map":
			instruct := iter.mappers[order.index]
			// no temp slice necessary
			for i := range workingSlice {
				workingSlice[i] = instruct(workingSlice[i])
			}

		case "take":
			takeUntilIndex := iter.takeCounts[order.index] - 1

			if takeUntilIndex > len(workingSlice)-1 {
				log.Printf("index %v out of range, skipping order...", takeUntilIndex)
				continue
			}

			tempSlice := make([]T, 0, len(workingSlice))
			for idx := 0; idx <= takeUntilIndex; idx++ {
				tempSlice = append(tempSlice, workingSlice[idx])
			}

			workingSlice = tempSlice

		case "skip":
			skipUntilIndex := iter.skipCounts[order.index] - 1

			if skipUntilIndex > len(workingSlice)-1 {
				log.Printf("index %v out of range. skipping order...", skipUntilIndex)
				continue
			}

			tempSlice := make([]T, 0, len(workingSlice)-(skipUntilIndex-1))
			for idx := skipUntilIndex + 1; idx < len(workingSlice); idx++ {
				tempSlice = append(tempSlice, workingSlice[idx])
			}

			workingSlice = tempSlice
		}
	}

	return workingSlice
}

func (iter Dei[T]) String() string {
	var out string

	for idx, val := range iter.orders {
		out += fmt.Sprintf(
			"Order %v:\n\tAdapter: %v\n\tIndex: %v\n\tComments %v\n",
			idx+1, val.method, val.index, val.comments,
		)
	}

	return out
}
