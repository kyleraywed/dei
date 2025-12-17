package derp

// Deferred Execution Reusable data-processing Pipeline

/*
	TODO:
		Apply() should return ([]T, error) and the logs from skip and apply should go there
*/

/*
	REMEMBER:
		Keep the methods in alphabetical order with Apply() & String() at the bottom.
		This applies to the switch in Apply() as well. Keep them in order.
*/

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"

	clone "github.com/huandu/go-clone/generic"
)

type order struct {
	method   string
	index    int
	comments []string
}

type Derp[T any] struct {
	filters    []func(t T) bool
	mappers    []func(t T) T
	foreachers []func(t T)

	takeCounts []int
	skipCounts []int

	orders []order
}

// Keep only the elements where in returns true. Optional comment strings.
func (pipeline *Derp[T]) Filter(in func(value T) bool, comments ...string) {
	pipeline.filters = append(pipeline.filters, in)
	pipeline.orders = append(pipeline.orders, order{
		method:   "filter",
		index:    len(pipeline.filters) - 1,
		comments: comments,
	})
}

// Perform logic using each element as an input. No changes to the underlying elements are made.
// Set the first optional comment to "con" for concurrent execution of input functions.
// Concurrent execution will be slower for most use-cases, while the order in which the funcs are
// evaluated is non-deterministic. Be careful when using "con".
func (pipeline *Derp[T]) Foreach(in func(value T), comments ...string) {
	pipeline.foreachers = append(pipeline.foreachers, in)
	pipeline.orders = append(pipeline.orders, order{
		method:   "foreach",
		index:    len(pipeline.foreachers) - 1,
		comments: comments,
	})
}

// Transform each value by applying a function. Optional comment strings.
func (pipeline *Derp[T]) Map(in func(value T) T, comments ...string) {
	pipeline.mappers = append(pipeline.mappers, in)
	pipeline.orders = append(pipeline.orders, order{
		method:   "map",
		index:    len(pipeline.mappers) - 1,
		comments: comments,
	})
}

// Skip the first n items and yields the rest. Comments inferred.
func (pipeline *Derp[T]) Skip(n int) error {
	if n < 1 {
		return fmt.Errorf("Skip(%v): No order submitted.", n)
	}

	pipeline.skipCounts = append(pipeline.skipCounts, n)
	pipeline.orders = append(pipeline.orders, order{
		method:   "skip",
		index:    len(pipeline.skipCounts) - 1,
		comments: []string{strconv.Itoa(n)},
	})

	return nil
}

// Yield only the first n items from the pipeline. Comments inferred.
func (pipeline *Derp[T]) Take(n int) error {
	if n < 1 {
		return fmt.Errorf("Take(%v): No order submitted.", n)
	}

	pipeline.takeCounts = append(pipeline.takeCounts, n)
	pipeline.orders = append(pipeline.orders, order{
		method:   "take",
		index:    len(pipeline.takeCounts) - 1,
		comments: []string{strconv.Itoa(n)},
	})

	return nil
}

// Interpret orders on data. Return new slice.
func (pipeline *Derp[T]) Apply(input []T) ([]T, error) {
	workingSlice := make([]T, len(input))
	workingSlice = clone.Clone(input) // deep clone by default

	numWorkers := runtime.GOMAXPROCS(0)
	// init chunksize
	chunkSize := (len(workingSlice) + numWorkers - 1) / numWorkers

	for _, order := range pipeline.orders {
		switch order.method {
		case "filter":
			workOrder := pipeline.filters[order.index]
			results := make([][]T, numWorkers)

			var wg sync.WaitGroup
			wg.Add(numWorkers)

			for idx := range numWorkers {
				start := idx * chunkSize

				if start >= len(workingSlice) {
					wg.Done()
					continue
				}

				// If the end marker runs longer than the slice, you've reached the end.
				end := min(start+chunkSize, len(workingSlice))

				chunk := workingSlice[start:end]

				go func() {
					defer wg.Done()

					out := make([]T, 0, len(chunk))
					for _, v := range chunk {
						if workOrder(v) {
							out = append(out, v)
						}
					}
					results[idx] = out
				}()
			}

			wg.Wait()

			// Flatten
			newlength := 0
			for _, r := range results {
				newlength += len(r)
			}
			tempSlice := make([]T, 0, newlength)

			for _, r := range results {
				tempSlice = append(tempSlice, r...)
			}

			workingSlice = tempSlice

		case "foreach":
			workOrder := pipeline.foreachers[order.index]

			if len(order.comments) > 0 && order.comments[0] == "con" {
				var wg sync.WaitGroup
				wg.Add(numWorkers)

				for idx := range numWorkers {
					start := idx * chunkSize

					if start >= len(workingSlice) {
						wg.Done()
						continue
					}

					end := min(start+chunkSize, len(workingSlice))

					chunk := workingSlice[start:end]

					go func() {
						defer wg.Done()

						for _, v := range chunk {
							workOrder(v)
						}
					}()
				}

				wg.Wait()

			} else {
				for _, val := range workingSlice {
					workOrder(val)
				}
			}

		case "map":
			workOrder := pipeline.mappers[order.index]

			var wg sync.WaitGroup
			wg.Add(numWorkers)

			for idx := range numWorkers {
				start := idx * chunkSize

				if start >= len(workingSlice) {
					wg.Done()
					continue
				}

				end := min(start+chunkSize, len(workingSlice))

				chunk := workingSlice[start:end]

				go func() {
					defer wg.Done()
					for i := range chunk {
						chunk[i] = workOrder(chunk[i])
					}
				}()
			}

			wg.Wait()

		case "skip":
			skipUntilIndex := pipeline.skipCounts[order.index]

			if skipUntilIndex > len(workingSlice) {
				return nil, fmt.Errorf("Skip(); index %v is out of range.", skipUntilIndex-1)
			}

			workingSlice = workingSlice[skipUntilIndex:]

		case "take":
			takeUntilIndex := pipeline.takeCounts[order.index]

			if takeUntilIndex > len(workingSlice) {
				return nil, fmt.Errorf("Take(); index %v is out of range", takeUntilIndex-1)
			}

			workingSlice = workingSlice[:takeUntilIndex]
		}
		// redistribute work evenly among workers after every order
		chunkSize = (len(workingSlice) + numWorkers - 1) / numWorkers
	}

	return workingSlice, nil
}

func (pipeline Derp[T]) String() string {
	var out strings.Builder

	for idx, val := range pipeline.orders {
		var prettyComments string

		if len(val.comments) == 0 {
			prettyComments += "[ N/A ]\n"
		}

		for _, cmt := range val.comments {
			prettyComments += "[ " + cmt + " ]\n\t\t"
		}

		fmt.Fprintf(&out, "Order %v:\n\tAdapter: %v\n\tIndex: %v\n\tComments: \n\t\t%v\n",
			idx+1, val.method, val.index, prettyComments)
	}

	return out.String()
}
