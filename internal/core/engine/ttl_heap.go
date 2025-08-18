package engine

type TtlHeapItem struct {
	Key      string
	ExpireAt int64
	Index    int
}

type TtlHeap []*TtlHeapItem

func (h TtlHeap) Len() int           { return len(h) }
func (h TtlHeap) Less(i, j int) bool { return h[i].ExpireAt < h[j].ExpireAt }
func (h TtlHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

func (h *TtlHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*TtlHeapItem)
	item.Index = n
	*h = append(*h, item)
}

func (h *TtlHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*h = old[0 : n-1]
	return item
}
