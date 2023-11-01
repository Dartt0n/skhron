package storage

import (
	"container/heap"
	"context"
	"log"
	"sync"
	"time"

	"errors"

	smap "github.com/go-auxiliaries/shrinking-map/pkg/shrinking-map"
)

type Storage struct {
	mu sync.RWMutex

	Data *smap.Map[string, []byte]
	TTLq *ExpQueue
}

// New function returns a new instance of the Storage struct
// with an initialized data map and a mutex.
func New() *Storage {
	storage := &Storage{
		mu: sync.RWMutex{},

		Data: smap.New[string, []byte](10000), // shrink map after every 10k deletions
		TTLq: NewExpQueue(),
	}
	heap.Init(storage.TTLq)
	return storage
}

// Put is a function which puts a value in the data map under a key.
// It takes the key as string and the value as byte slice.
// This function locks write & read mutex for its operations.
func (s *Storage) Put(key string, value []byte, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data.Set(key, value)

	// Find previous item
	index := -1
	for i := 0; i < s.TTLq.Len(); i++ {
		if (*s.TTLq)[i].key == key {
			index = i
			(*s.TTLq)[index].exp = time.Now().Add(ttl)
		}
	}

	if index == -1 {
		heap.Push(s.TTLq, &ItemTTL{key: key, exp: time.Now().Add(ttl)})
	} else {
		heap.Fix(s.TTLq, index)
	}

	return nil
}

// Get is a function which fetches a value in the data map under a key.
// It takes the key as string parameter.
// If the key is not present, error is returned.
// This function locks read mutex for its operations.
func (s *Storage) Get(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if v, ok := s.Data.Get2(key); ok {
		return v, nil
	}

	return []byte{}, errors.New("no such key: " + key)
}

// Delete is a function which deletes a key from the data map.
// It takes the key as string parameter.
// This function locks read & write mutex for its operations.
func (s *Storage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data.Delete(key)

	return nil
}

// Exists is a function which check wheater a key is present in the data map.
// It takes the key as string parameter.
// This function locks read mutex for its operations.
func (s *Storage) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exist := s.Data.Get2(key)
	return exist
}

// CleanUp is a function which removes expired items.
// It is called periodically by the `CleaningProcess` function.
func (s *Storage) CleanUp() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Storage cleanup started\n")

	now := time.Now()
	deleted := 0

	for s.TTLq.Len() > 0 {
		item := heap.Pop(s.TTLq).(*ItemTTL)

		if item.exp.Before(now) {
			log.Printf("Item with key \"%s\" expired %f sec ago, deleting\n", item.key, now.Sub(item.exp).Seconds())
			s.Data.Delete(item.key)
			deleted++
		} else {
			heap.Push(s.TTLq, item)
			break
		}
	}

	log.Printf("Storage cleanup finished. %d keys deleted, %d left in queue\n", deleted, s.TTLq.Len())
}

// The `CleaningProcess` function is a goroutine that runs
// in the background and periodically calls the
// `CleanUp` function of the `Storage` struct.
// It works until ctx.Done() is called.
// It puts into done channel when it finishes.
// It runs clean up process every `period` time duration.
func (s *Storage) CleaningProcess(ctx context.Context, period time.Duration, done chan struct{}) {
	log.Printf("Starting cleaning up process with period %.02f sec\n", period.Seconds())
loop:
	for {
		select {
		case <-ctx.Done():
			s.CleanUp()
			log.Println("Shutting down storage cleanup process")
			break loop
		case <-time.After(period):
			s.CleanUp()
		}
	}

	done <- struct{}{}
}
