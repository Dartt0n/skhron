package storage

import "time"

type ItemTTL struct {
	Key string    `json:"key,omitempty"`
	Exp time.Time `json:"exp,omitempty"`
}

type ExpireQueue []*ItemTTL

func NewExpQueue() *ExpireQueue {
	x := make(ExpireQueue, 0)
	return &x
}

// heap.Interface
func (q ExpireQueue) Len() int           { return len(q) }
func (q ExpireQueue) Less(i, j int) bool { return q[i].Exp.Before(q[j].Exp) }
func (q ExpireQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }

func (q *ExpireQueue) Push(x any) {
	*q = append(*q, x.(*ItemTTL))
}

func (q *ExpireQueue) Pop() any {
	qcopy := *q
	n := len(qcopy)
	item := qcopy[n-1]
	qcopy[n-1] = nil // avoid memory leak

	*q = qcopy[0 : n-1]
	return item
}
