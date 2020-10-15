package simplescheduler

import (
	"container/heap"
	"testing"
)

func generateData() []*schedHeapItem {

	result := []*schedHeapItem{
		&schedHeapItem{
			values:   []string{"A"},
			priority: 30,
		},

		&schedHeapItem{
			values:   []string{"B"},
			priority: 20,
		},

		&schedHeapItem{
			values:   []string{"C"},
			priority: 10,
		},
	}

	return result
}

func TestLen(t *testing.T) {
	h := make(SchedMinHeap, 0)
	data := generateData()
	data_n := len(data)

	heap.Init(&h)

	for _, x := range data {
		heap.Push(&h, x)
	}

	if len(h) != data_n {
		t.Errorf("len(h) = %d; wants %d", len(h), data_n)
	}

	if h.Len() != len(h) {
		t.Errorf("h.Len() = %d; wants %d", h.Len(), len(h))
	}
}

func TestPush(t *testing.T) {
	h := make(SchedMinHeap, 0)
	heap.Init(&h)
	data := generateData()

	heap.Push(&h, data[0])
	heap.Push(&h, data[1])
	heap.Push(&h, data[2])

	n := 3
	if h.Len() != n {
		t.Errorf("len(h) = %d; wants %d", len(h), n)
	}

	table := []int64{10, 30, 20}

	for i, x := range table {
		if h[i].priority != x {
			t.Errorf("h[i].priority = %d; wants %d", h[i].priority, x)
		}
	}

}
