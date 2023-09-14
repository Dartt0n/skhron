package main

import (
	"container/heap"
	"context"
	"log"
	"sync"
	"time"

	"errors"
)

type Storage struct {
	_datamu sync.Mutex
	data    map[string][]byte

	_ttlmu sync.Mutex
	ttl    heap.Interface
}

// NewStorage function returns a new instance of the Storage struct
// with an initialized data map and a mutex.
func NewStorage() *Storage {
	s := &Storage{
		_datamu: sync.Mutex{},
		data:    make(map[string][]byte),

		_ttlmu: sync.Mutex{},
		ttl:    NewExpQueue(),
	}
	heap.Init(s.ttl)
	return s
}

// Put is a function which puts a value in the data map under a key.
// It takes the key as string and the value as byte slice.
// This function locks _datamu mutex for its operations.
func (s *Storage) Put(key string, value []byte, ttl time.Duration) error {
	s._datamu.Lock()
	defer s._datamu.Unlock()
	s.data[key] = value

	s._ttlmu.Lock()
	defer s._ttlmu.Unlock()
	heap.Push(s.ttl, &ItemTTL{key: key, exp: time.Now().Add(ttl)})

	return nil
}

// Get is a function which fetches a value in the data map under a key.
// It takes the key as string parameter.
// If the key is not present, error is returned.
// This function locks _datamu mutex for its operations.
func (s *Storage) Get(key string) ([]byte, error) {
	s._datamu.Lock()
	defer s._datamu.Unlock()

	if v, ok := s.data[key]; ok {
		return v, nil
	}

	return []byte{}, errors.New("no such key: " + key)
}

// Delete is a function which deletes a key from the data map.
// It takes the key as string parameter.
// This function locks _datamu mutex for its operations.
func (s *Storage) Delete(key string) error {
	s._datamu.Lock()
	defer s._datamu.Unlock()

	delete(s.data, key)

	return nil
}

// Exists is a function which check wheater a key is present in the data map.
// It takes the key as string parameter.
// This function locks _datamu mutex for its operations.
func (s *Storage) Exists(key string) bool {
	s._datamu.Lock()
	defer s._datamu.Unlock()

	_, exist := s.data[key]
	return exist
}

// CleanUp is a function which removes expired items.
// It is called periodically by the `CleaningProcess` function.
func (s *Storage) CleanUp() {
	s._ttlmu.Lock()
	defer s._ttlmu.Unlock()
	log.Printf("Storage cleanup started\n")

	now := time.Now()

	for s.ttl.Len() > 0 {
		item := heap.Pop(s.ttl).(*ItemTTL)

		if item.exp.Before(now) {
			log.Printf("Item with key \"%s\" expired %f sec ago, deleting\n", item.key, now.Sub(item.exp).Seconds())
			s.Delete(item.key)
		} else {
			heap.Push(s.ttl, item)
			break
		}
	}

	log.Printf("Storage cleanup finished\n")
}

// The `CleaningProcess` function is a goroutine that runs
// in the background and periodically calls the
// `CleanUp` function of the `Storage` struct.
func (s *Storage) CleaningProcess(ctx context.Context, period time.Duration, done chan struct{}) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			s.CleanUp()
			time.Sleep(period)
		}
	}

	log.Println("Storage cleanup process finished")
	done <- struct{}{}
}
