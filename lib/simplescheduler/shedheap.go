package simplescheduler

// An item of min-heap
type schedHeapItem struct {
	value    string
	priority int64
}

type SchedMinHeap []*schedHeapItem

func (h SchedMinHeap) Len() int {

	return len(h)
}

func (h SchedMinHeap) Less(i, j int) bool {

	return h[i].priority < h[j].priority
}

func (h SchedMinHeap) Swap(i, j int) {

	h[i], h[j] = h[j], h[i]
}

func (h *SchedMinHeap) Push(x interface{}) {
	*h = append(*h, x.(*schedHeapItem))
}

func (h *SchedMinHeap) Pop() interface{} {

	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*h = old[0 : n-1]
	return item
}
