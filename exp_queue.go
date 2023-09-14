package main

import "time"

type ItemTTL struct {
	key string
	exp time.Time
}

type ExpQueue []*ItemTTL

func NewExpQueue() *ExpQueue {
	x := make(ExpQueue, 0)
	return &x
}

// heap.Interface
func (q ExpQueue) Len() int           { return len(q) }
func (q ExpQueue) Less(i, j int) bool { return q[i].exp.Before(q[j].exp) }
func (q ExpQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }

func (q *ExpQueue) Push(x any) {
	*q = append(*q, x.(*ItemTTL))
}

func (q *ExpQueue) Pop() any {
	qcopy := *q
	n := len(qcopy)
	item := qcopy[n-1]
	qcopy[n-1] = nil // avoid memory leak

	*q = qcopy[0 : n-1]
	return item
}
