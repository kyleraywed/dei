package dei

import (
	"fmt"
	"log"
	"strconv"
)

type order struct {
	adapter  string
	index    int
	comments []string
}

type Dei[T any] struct {
	filterInstruct []func(t T) bool
	mapInstruct    []func(t T) T
	takeIndexes    []int
	skipIndexes    []int

	orders []order
}

// Keep only the elements where in returns true. Optional comment strings.
func (iter *Dei[T]) Filter(in func(value T) bool, comments ...string) {
	iter.filterInstruct = append(iter.filterInstruct, in)
	iter.orders = append(iter.orders, order{
		adapter: "filter", index: len(iter.filterInstruct) - 1, comments: comments,
	})
}

// Transform each element by applying a function. Optional comment strings.
func (iter *Dei[T]) Map(in func(value T) T, comments ...string) {
	iter.mapInstruct = append(iter.mapInstruct, in)
	iter.orders = append(iter.orders, order{
		adapter: "map", index: len(iter.mapInstruct) - 1, comments: comments,
	})
}

// Yield only the first n items from the iterator. Comments inferred.
func (iter *Dei[T]) Take(n int) {
	if n < 1 {
		log.Printf("Take(%v): No order submitted.", n)
		return
	}

	iter.takeIndexes = append(iter.takeIndexes, n-1)

	iter.orders = append(iter.orders, order{
		adapter: "take", index: len(iter.takeIndexes) - 1, comments: []string{strconv.Itoa(n)},
	})
}

// Skip the first n items and yields the rest. Comments inferred.
func (iter *Dei[T]) Skip(n int) {
	if n < 1 {
		log.Printf("Skip(%v): No order submitted.", n)
		return
	}

	iter.skipIndexes = append(iter.skipIndexes, n-1)

	iter.orders = append(iter.orders, order{
		adapter: "skip", index: len(iter.skipIndexes) - 1, comments: []string{strconv.Itoa(n)},
	})
}

// Interpret orders on data. Return new slice.
func (iter *Dei[T]) Apply(input []T) []T {
	workingSlice := input

	for _, order := range iter.orders {
		switch order.adapter {

		case "filter":
			tempSlice := make([]T, 0, len(workingSlice))
			instruct := iter.filterInstruct[order.index]

			for _, val := range workingSlice {
				if instruct(val) {
					tempSlice = append(tempSlice, val)
				}
			}

			workingSlice = tempSlice

		case "map":
			instruct := iter.mapInstruct[order.index]
			// no temp slice necessary
			for i := range workingSlice {
				workingSlice[i] = instruct(workingSlice[i])
			}

		case "take":
			takeIndex := iter.takeIndexes[order.index]

			if takeIndex > len(workingSlice)-1 {
				log.Printf("index %v out of range, skipping order...", takeIndex)
				continue
			}

			tempSlice := make([]T, 0, len(workingSlice))
			for idx := 0; idx <= takeIndex; idx++ {
				tempSlice = append(tempSlice, workingSlice[idx])
			}

			workingSlice = tempSlice

		case "skip":
			skipIndex := iter.skipIndexes[order.index]

			if skipIndex > len(workingSlice)-1 {
				log.Printf("index %v out of range. skipping order...", skipIndex)
				continue
			}

			tempSlice := make([]T, 0, len(workingSlice)-(skipIndex-1))
			for idx := skipIndex + 1; idx < len(workingSlice); idx++ {
				tempSlice = append(tempSlice, workingSlice[idx])
			}

			workingSlice = tempSlice
		}
	}

	return workingSlice
}

func (iter Dei[T]) String() string {
	out := fmt.Sprintf(
		"Filter instruction addresses:\n\t%v\nMap instruction addresses:\n\t%v\n\n",
		iter.filterInstruct, iter.mapInstruct,
	)

	for idx, val := range iter.orders {
		out += fmt.Sprintf(
			"Order %v:\n\tAdapter: %v\n\tIndex: %v\n\tComments %v\n",
			idx+1, val.adapter, val.index, val.comments,
		)
	}

	return out
}
