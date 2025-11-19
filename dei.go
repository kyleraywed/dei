package dei

import (
	"fmt"
	"log"
	"strconv"
)

type Dei[T any] struct {
	filterInstruct []func(t T) bool
	mapInstruct    []func(t T) T
	takeIndexes    []int
	skipIndexes    []int

	orders []struct {
		adapter  string   // filter, map, etc
		index    int      // the index of the slice in which the instruct or index needed is held
		comments []string // for debug printing
	}
}

// Keeps only the elements where in returns true. Optional comment strings.
func (iter *Dei[T]) Filter(in func(value T) bool, comments ...string) {
	iter.filterInstruct = append(iter.filterInstruct, in)
	iter.orders = append(iter.orders, struct {
		adapter  string
		index    int
		comments []string
	}{
		adapter: "filter", index: len(iter.filterInstruct) - 1, comments: comments,
	})
}

// Transforms each element by applying a function. Optional comment strings.
func (iter *Dei[T]) Map(in func(value T) T, comments ...string) {
	iter.mapInstruct = append(iter.mapInstruct, in)
	iter.orders = append(iter.orders, struct {
		adapter  string
		index    int
		comments []string
	}{
		adapter: "map", index: len(iter.mapInstruct) - 1, comments: comments,
	})
}

// Yields only the first n items from the iterator. Comments inferred.
func (iter *Dei[T]) Take(n int) {
	if n < 1 {
		log.Printf("Take(%v): No order submitted.", n)
		return
	}

	iter.takeIndexes = append(iter.takeIndexes, n-1)

	iter.orders = append(iter.orders, struct {
		adapter  string
		index    int
		comments []string
	}{
		adapter: "take", index: len(iter.takeIndexes) - 1, comments: []string{strconv.Itoa(n)},
	})
}

// Skips the first n items and yields the rest. Comments inferred.
func (iter *Dei[T]) Skip(n int) {
	if n < 1 {
		log.Printf("Skip(%v): No order submitted.", n)
		return
	}

	iter.skipIndexes = append(iter.skipIndexes, n-1)

	iter.orders = append(iter.orders, struct {
		adapter  string
		index    int
		comments []string
	}{
		adapter: "skip", index: len(iter.skipIndexes) - 1, comments: []string{strconv.Itoa(n)},
	})
}

// Interpret orders on data. Return new slice.
func (iter *Dei[T]) Apply(in []T) []T {
	out := in

	for _, order := range iter.orders {
		switch order.adapter {

		case "filter":
			new := make([]T, 0, len(out))
			instruct := iter.filterInstruct[order.index]

			for _, val := range out {
				if instruct(val) {
					new = append(new, val)
				}
			}

			out = new

		case "map":
			instruct := iter.mapInstruct[order.index]
			for i := range out {
				out[i] = instruct(out[i])
			}

		case "take":
			index := iter.takeIndexes[order.index]

			if index > len(out)-1 {
				log.Printf("index %v out of range, skipping order...", index)
				continue
			}

			new := make([]T, 0, len(out))
			for idx := 0; idx <= index; idx++ {
				new = append(new, out[idx])
			}

			out = new

		case "skip":
			index := iter.skipIndexes[order.index]

			if index > len(out)-1 {
				log.Printf("index %v out of range. skipping order...", index)
				continue
			}

			new := make([]T, 0, len(out)-(index-1))
			for idx := index + 1; idx < len(out); idx++ {
				new = append(new, out[idx])
			}

			out = new
		}
	}

	return out
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
