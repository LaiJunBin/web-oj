package sharedtypes

import (
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHeap_New(t *testing.T) {
	t.Run("should create a new heap with min heap", func(t *testing.T) {
		h := NewHeap([]int{3, 1, 2}, func(i, j int) bool {
			return i < j
		})

		assert.Equal(t, 3, h.Len())

		expected := []int{1, 2, 3}
		actual := []int{}
		for h.Len() > 0 {
			actual = append(actual, h.Pop())
		}

		assert.Equal(t, expected, actual)
	})

	t.Run("should create a new heap with max heap", func(t *testing.T) {
		h := NewHeap([]int{3, 1, 2}, func(i, j int) bool {
			return i > j
		})

		assert.Equal(t, 3, h.Len())

		expected := []int{3, 2, 1}
		actual := []int{}
		for h.Len() > 0 {
			actual = append(actual, h.Pop())
		}

		assert.Equal(t, expected, actual)
	})

	t.Run("should create a new heap with min heap by priority", func(t *testing.T) {
		type Item struct {
			value    string
			priority int
		}

		items := []Item{
			{"a", 3},
			{"b", 1},
			{"c", 2},
		}

		h := NewHeap(items, func(i, j Item) bool {
			return i.priority < j.priority
		})

		assert.Equal(t, 3, h.Len())

		expected := []string{"b", "c", "a"}
		actual := []string{}
		for h.Len() > 0 {
			actual = append(actual, h.Pop().value)
		}

		assert.Equal(t, expected, actual)
	})
}

func TestHeap_Push(t *testing.T) {
	t.Run("should push new item to heap", func(t *testing.T) {
		h := NewHeap([]int{3, 1, 2}, func(i, j int) bool {
			return i < j
		})

		h.Push(4)
		h.Push(0)

		assert.Equal(t, 5, h.Len())

		expected := []int{0, 1, 2, 3, 4}
		actual := []int{}
		for h.Len() > 0 {
			actual = append(actual, h.Pop())
		}

		assert.Equal(t, expected, actual)
	})

	t.Run("should push new item to heap by priority", func(t *testing.T) {
		type Item struct {
			value    string
			priority int
		}

		items := []Item{
			{"a", 3},
			{"b", 1},
			{"c", 2},
		}

		h := NewHeap(items, func(i, j Item) bool {
			return i.priority < j.priority
		})

		h.Push(Item{"d", 0})
		h.Push(Item{"e", 4})

		assert.Equal(t, 5, h.Len())

		expected := []string{"d", "b", "c", "a", "e"}
		actual := []string{}
		for h.Len() > 0 {
			actual = append(actual, h.Pop().value)
		}

		assert.Equal(t, expected, actual)
	})
}

func TestHeap_PopEmptyDoesNotBlock(t *testing.T) {
	h := NewHeap([]int{}, func(i, j int) bool {
		return i < j
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = h.Pop()
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Pop on empty heap should not block")
	}
}

func TestHeap_PeekEmptyDoesNotBlock(t *testing.T) {
	h := NewHeap([]int{}, func(i, j int) bool {
		return i < j
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = h.Peek()
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Peek on empty heap should not block")
	}
}

func TestHeap_ConcurrentPushPop(t *testing.T) {
	h := NewHeap([]int{}, func(i, j int) bool {
		return i < j
	})

	const (
		producers    = 4
		perProducer  = 250
		totalEntries = producers * perProducer
	)

	var producersWG sync.WaitGroup
	producersWG.Add(producers)

	for p := 0; p < producers; p++ {
		go func(offset int) {
			defer producersWG.Done()
			base := offset*perProducer + 1
			for i := 0; i < perProducer; i++ {
				h.Push(base + i)
			}
		}(p)
	}

	resultsCh := make(chan []int, 1)
	go func() {
		collected := make([]int, 0, totalEntries)
		for len(collected) < totalEntries {
			val := h.Pop()
			if val == 0 {
				time.Sleep(100 * time.Microsecond)
				continue
			}
			collected = append(collected, val)
		}
		resultsCh <- collected
	}()

	producersWG.Wait()

	var results []int

	select {
	case results = <-resultsCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for heap results")
	}

	assert.Equal(t, totalEntries, len(results), "should pop exactly the number of pushed items")

	sortedResults := make([]int, len(results))
	copy(sortedResults, results)
	sort.Ints(sortedResults)

	expected := make([]int, totalEntries)
	for i := range expected {
		expected[i] = i + 1
	}

	assert.Equal(t, expected, sortedResults)
}
