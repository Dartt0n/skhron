package storage

import (
	"container/heap"
	"context"
	"encoding/json"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"errors"

	smap "github.com/go-auxiliaries/shrinking-map/pkg/shrinking-map"
)

type Storage struct {
	mu sync.RWMutex

	Data *smap.Map[string, []byte] `json:"data,omitempty"`
	TTLq *ExpireQueue              `json:"ttlq,omitempty"`
}

// New function returns a new instance of the Storage struct.
func New() *Storage {
	storage := &Storage{
		mu: sync.RWMutex{},

		Data: smap.New[string, []byte](10000), // shrink map after every 10k deletions
		TTLq: NewExpQueue(),
	}
	heap.Init(storage.TTLq)
	return storage
}

// Put is a function which puts a value in the storage under a key.
// It takes the key as string and the value as byte slice.
// This function locks mutex for its operations.
func (s *Storage) Put(key string, value []byte, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data.Set(key, value)

	// Find previous item
	index := -1
	for i := 0; i < s.TTLq.Len(); i++ {
		if (*s.TTLq)[i].Key == key {
			index = i
			(*s.TTLq)[index].Exp = time.Now().Add(ttl)
		}
	}

	if index == -1 {
		heap.Push(s.TTLq, &ItemTTL{Key: key, Exp: time.Now().Add(ttl)})
	} else {
		heap.Fix(s.TTLq, index)
	}

	return nil
}

// Get is a function which fetches a value in the storage under a key.
// It takes the key as string parameter.
// If the key is not present, error is returned.
// This function locks mutex for its operations.
func (s *Storage) Get(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if v, ok := s.Data.Get2(key); ok {
		return v, nil
	}

	return []byte{}, errors.New("no such key: " + key)
}

// Delete is a function which deletes a key from the storage.
// It takes the key as string parameter.
// This function locks mutex for its operations.
func (s *Storage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data.Delete(key)

	return nil
}

// Exists is a function which check wheater a key is present in the storage.
// It takes the key as string parameter.
// This function locks mutex for its operations.
func (s *Storage) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exist := s.Data.Get2(key)
	return exist
}

// CleanUp is a function which removes expired items.
// It is called periodically by the `PeriodicCleanup` function.
// This function locks mutex for its operations.
func (s *Storage) CleanUp() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Storage cleanup started\n")

	now := time.Now()
	deleted := 0

	for s.TTLq.Len() > 0 {
		item := heap.Pop(s.TTLq).(*ItemTTL)

		if item.Exp.Before(now) {
			log.Printf("Item with key \"%s\" expired %f sec ago, deleting\n", item.Key, now.Sub(item.Exp).Seconds())
			s.Data.Delete(item.Key)
			deleted++
		} else {
			heap.Push(s.TTLq, item)
			break
		}
	}

	log.Printf("Storage cleanup finished. %d keys deleted, %d left in queue\n", deleted, s.TTLq.Len())
}

// `PeriodicCleanup` is a function that
// periodically calls the `CleanUp` function.
// It works until `ctx.Done()` signal is sent.
// It puts into `done` channel when it finishes.
// It backups current state of the storage into file `./skhron/skhron_{timestamp}.json` on exit.
// It runs clean up process every `period` time duration.
func (s *Storage) PeriodicCleanup(ctx context.Context, period time.Duration, done chan struct{}) {
	log.Printf("Starting cleaning up process with period %.02f sec\n", period.Seconds())
loop:
	for {
		select {
		case <-ctx.Done():
			s.CleanUp()

			// todo: logging
			if err := s.FileBackup(time.Now().Format("skhron_2006_01_02_15:04:05.json")); err != nil {
				log.Printf("Failed to backup file: %v\n", err)
			}

			log.Println("Shutting down storage cleanup process")
			break loop
		case <-time.After(period):
			s.CleanUp()
		}
	}

	done <- struct{}{}
}

func (s *Storage) FileBackup(file string) error { // todo: better method names
	s.mu.Lock()
	defer s.mu.Unlock()

	bytes, err := json.Marshal(map[string]interface{}{
		"data": s.Data.Values(),
		"ttlq": s.TTLq,
	})

	if err != nil {
		return err
	}

	if dir, err := os.Stat(".skhron"); os.IsNotExist(err) {
		err = os.Mkdir(".skhron", 0777)
		if err != nil {
			return err
		}
	} else if !dir.IsDir() {
		return errors.New(".skhron file is not a directory")
	}

	f, err := os.Create(path.Join(".skhron", file))
	if err != nil {
		return err
	}

	_, err = f.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) FileLoad(path string) error {
	// TODO: find last file
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Open(path)
	if err != nil {
		return err
	}

	rs := &struct {
		Data map[string][]byte
		TTLq *ExpireQueue
	}{}

	dec := json.NewDecoder(f)
	if err := dec.Decode(rs); err != nil {
		return err
	}

	for rs.TTLq.Len() > 0 {
		item := rs.TTLq.Pop().(*ItemTTL)
		s.Data.Set(item.Key, rs.Data[item.Key])
		s.TTLq.Push(item)
	}

	return nil
}
