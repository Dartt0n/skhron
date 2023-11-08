package skhron

import "time"

type expireItem struct {
	Key string    `json:"key,omitempty"`
	Exp time.Time `json:"exp,omitempty"`
}

type expireQueue []*expireItem

func newExpQueue() *expireQueue {
	x := make(expireQueue, 0)
	return &x
}

// heap.Interface
func (q expireQueue) Len() int           { return len(q) }
func (q expireQueue) Less(i, j int) bool { return q[i].Exp.Before(q[j].Exp) }
func (q expireQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }

func (q *expireQueue) Push(x any) {
	*q = append(*q, x.(*expireItem))
}

func (q *expireQueue) Pop() any {
	qcopy := *q
	n := len(qcopy)
	item := qcopy[n-1]
	qcopy[n-1] = nil // avoid memory leak

	*q = qcopy[0 : n-1]
	return item
}
