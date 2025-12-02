package dei

import (
	"slices"
	"strconv"
	"sync"
	"testing"
)

func TestFilter(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var iter Dei[int]

	iter.Filter(func(value int) bool {
		return value%2 == 0 // return evens
	})

	expected := []int{2, 4, 6, 8, 10}
	gotten := iter.Apply(numbers)

	if len(expected) != len(gotten) {
		t.Error("Filter len mismatch")
	}

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("Filter value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestMap(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var iter Dei[int]

	iter.Map(func(value int) int {
		return value * value // square the numbers
	})

	expected := []int{1, 4, 9, 16, 25, 36, 49, 64, 81, 100}
	gotten := iter.Apply(numbers)

	if len(expected) != len(gotten) {
		t.Error("Map len mismatch")
	}

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("Map value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestTake(t *testing.T) {
	numbers := []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21}
	var iter Dei[int]

	iter.Take(11)

	expected := []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21}
	gotten := iter.Apply(numbers)

	if len(expected) != len(gotten) {
		t.Error("Take len mismatch")
	}

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("Take value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestSkip(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var iter Dei[int]

	iter.Skip(3)

	expected := []int{4, 5, 6, 7, 8, 9, 10}
	gotten := iter.Apply(numbers)

	if len(expected) != len(gotten) {
		t.Error("Skip len mismatch")
	}

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("Skip value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestForeach(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var iter Dei[int]

	expected := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	var gotten []string

	iter.Foreach(func(value int) {
		gotten = append(gotten, strconv.Itoa(value))
	})

	iter.Apply(numbers)

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("Foreach value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestForeachFast(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var iter Dei[int]

	expected := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	var gotten []string

	var mu sync.Mutex

	iter.Foreach(func(value int) {
		mu.Lock()
		gotten = append(gotten, strconv.Itoa(value))
		mu.Unlock()
	}, "fast")

	iter.Apply(numbers)

	slices.SortFunc(gotten, func(a, b string) int {
		ai, _ := strconv.Atoi(a)
		bi, _ := strconv.Atoi(b)
		return ai - bi
	})

	for idx, val := range expected {
		if gotten[idx] != val {
			t.Errorf("Foreach value mismatch.\nExpected: [%v] Got: [%v]\n", expected, gotten)
		}
	}
}

func TestOrder(t *testing.T) {
	var iter Dei[int]

	iter.Filter(func(value int) bool {
		return value%2 == 0
	}, "Foo")

	iter.Map(func(value int) int {
		return value * 2
	}, "Bar")

	iter.Take(3)

	iter.Skip(1)

	iter.Map(func(value int) int {
		return value + 1
	}, "baz")

	iter.Filter(func(value int) bool {
		return value%2 != 0
	}, "boo")

	iter.Take(3)

	expected := []struct {
		adapter  string
		index    int
		comments []string
	}{
		{adapter: "filter", index: 0, comments: []string{"Foo"}},
		{adapter: "map", index: 0, comments: []string{"Bar"}},
		{adapter: "take", index: 0, comments: []string{"3"}},
		{adapter: "skip", index: 0, comments: []string{"1"}},
		{adapter: "map", index: 1, comments: []string{"baz"}},
		{adapter: "filter", index: 1, comments: []string{"boo"}},
		{adapter: "take", index: 1, comments: []string{"3"}},
	}

	if len(iter.orders) != len(expected) {
		t.Error("Order len mismatch")
	}

	for idx, val := range expected {
		if iter.orders[idx].method != val.adapter {
			t.Errorf("Order adapter mismatch.\nExpected [%v] Got: [%v]\n", val.adapter, iter.orders[idx].method)
		}
		if iter.orders[idx].index != val.index {
			t.Errorf("Order index mismatch.\nExpected [%v] Got: [%v]\n", val.index, iter.orders[idx].index)
		}
		if iter.orders[idx].comments[0] != val.comments[0] {
			t.Errorf("Order comment mismatch.\nExpected [%v] Got: [%v]\n", val.comments, iter.orders[idx].comments[0])
		}
	}

	//fmt.Println(iter)
}
