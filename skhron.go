package skhron

import (
	"container/heap"
	"context"
	"encoding/json"
	"log"
	"os"
	"path"
	"regexp"
	"sync"
	"time"

	"errors"

	smap "github.com/go-auxiliaries/shrinking-map/pkg/shrinking-map"
)

type Skhron[V any] struct { // todo: serializable/deserializable type
	mu sync.RWMutex // skhron data map mutex

	// Skhron.Data is a main object. All the data is stored here.
	// It is shrinking map, which shrinks every "limit" (skhron.WithMapLimit option) deletions.
	Data *smap.Map[string, V] `json:"data,omitempty"`
	// Skhron.TTLq is a queue object used to delete expired items in time.
	// Each put operation the object is either added to the queue or updated in the queue.
	// The cleaning up process takes an item from the front of the queue and checks,
	// Whether the item has expired or not. Thus, we avoid scanning the whole map to find expired items.
	TTLq *expireQueue `json:"ttlq,omitempty"`

	// Config

	// A directory where snapshots would be stored
	SnapshotDir string
	// A name (WITHOUT EXTENSION) which would be used to store the latest snapshot
	SnapshotName string
	// A directory where temporary files would be stored
	TempSnapshotDir string
}

// Initialize Skhron instance with options.
func New[V any](opts ...StorageOpt[V]) *Skhron[V] {
	skhron := &Skhron[V]{
		mu: sync.RWMutex{},

		Data: smap.New[string, V](0),
		TTLq: newExpQueue(),
	}

	heap.Init(skhron.TTLq) // initialize queue

	DefaultOpts(skhron)        // default options
	for _, opt := range opts { // iterate over provided options
		opt(skhron) // apply provided option
	}

	return skhron
}

// Put is a function which puts a value in the storage under a key.
// It takes the key as string and the value as V.
// This function locks mutex for its operations.
func (s *Skhron[V]) Put(key string, value V) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data.Set(key, value)

	return nil
}

// PutTTL is a function which puts a value in the storage under a key with certain TTL.
// It takes the key as string, the value as V and ttl as time.Duration.
// It scans the entire queue to find if the item is already in the queue.
// If it is, it updates the item and updates the queue (to maintain priority).
// If it is not, it puts the item into the queue.
// This function locks mutex for its operations.
func (s *Skhron[V]) PutTTL(key string, value V, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data.Set(key, value)

	// Find previous item
	index := -1
	for i := 0; i < s.TTLq.Len(); i++ { // iterate over the entire queue
		if (*s.TTLq)[i].Key == key {
			index = i
			(*s.TTLq)[index].Exp = time.Now().Add(ttl) // update item ttl
		}
	}

	if index == -1 {
		heap.Push(s.TTLq, &expireItem{Key: key, Exp: time.Now().Add(ttl)})
	} else {
		heap.Fix(s.TTLq, index) // update queue
	}

	return nil
}

// Get is a function which fetches a value in the storage under a key.
// It takes the key as string parameter.
// If the key is not present, error is returned.
// This function locks mutex for its operations.
func (s *Skhron[V]) Get(key string) (V, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if v, ok := s.Data.Get2(key); ok {
		return v, nil
	}

	return *new(V), errors.New("no such key: " + key)
}

func (s *Skhron[V]) GetRegex(mask *regexp.Regexp) []V {
	s.mu.Lock()
	defer s.mu.Unlock()

	data := s.Data.Values()

	values := make([]V, 0, len(data))

	for key, value := range data {
		if mask.Match([]byte(key)) {
			values = append(values, value)
		}
	}

	return values
}

// Delete is a function which deletes a key from the storage.
// It takes the key as string parameter.
// It does not delete the key from the queue, since when the key is expired
// it would be deleted from the queue without side effects.
// This function locks mutex for its operations.
func (s *Skhron[V]) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data.Delete(key)

	// todo: delete from queue?
	// benefit: lower memory usage
	// drawback: higher execution time

	return nil
}

// Exists is a function which check wheater a key is present in the storage.
// It takes the key as string parameter.
// This function locks mutex for its operations.
func (s *Skhron[V]) Exists(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exist := s.Data.Get2(key)
	return exist
}

// CleanUp is a function which removes expired items.
// It is called periodically by the `PeriodicCleanup` function.
// This function locks mutex for its operations.
func (s *Skhron[V]) CleanUp() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Skhron cleanup started\n")

	now := time.Now()
	deleted := 0

	for s.TTLq.Len() > 0 {
		item := heap.Pop(s.TTLq).(*expireItem)

		if item.Exp.Before(now) {
			log.Printf("Item with key \"%s\" expired %f sec ago, deleting\n", item.Key, now.Sub(item.Exp).Seconds())
			s.Data.Delete(item.Key)
			deleted++
		} else {
			heap.Push(s.TTLq, item)
			break
		}
	}

	log.Printf("Skhron cleanup finished. %d keys deleted, %d left in queue\n", deleted, s.TTLq.Len())
}

// `PeriodicCleanup` is a function that
// periodically calls the `CleanUp` function.
// It works until `ctx.Done()` signal is sent.
// It puts into `done` channel when it finishes.
// It backups current state of the storage into file `./skhron/skhron_{timestamp}.json` on exit.
// It runs clean up process every `period` time duration.
func (s *Skhron[V]) PeriodicCleanup(ctx context.Context, period time.Duration, done chan struct{}) {
	log.Printf("Starting cleaning up process with period %.02f sec\n", period.Seconds())
loop:
	for {
		select {
		case <-ctx.Done():
			s.CleanUp()

			if err := s.CreateSnapshot(); err != nil {
				log.Printf("failed to create snapshot file: %v\n", err)
			}

			log.Println("Shutting down skhron cleanup process")
			break loop
		case <-time.After(period):
			s.CleanUp()
		}
	}

	done <- struct{}{}
}

// JsonMarshal is a function, which converts the struct into JSON-string bytes
func (s *Skhron[V]) MarshalJSON() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bytes, err := json.Marshal(map[string]interface{}{
		"data": s.Data.Values(),
		"ttlq": s.TTLq,
	})

	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

// CreateSnapshot is a function which create snapshot (json dump of struct)
// in the temporary directory.
// Then it checks if an older snapshot exists in snapshot directory.
// If it is, it renames it to format "{snapshot name}_{time stamp}.skh"
// and then moves new snapshot to the snapshot directory
func (s *Skhron[V]) CreateSnapshot() error {
	// Marshal stroge to json
	bytes, err := s.MarshalJSON()
	if err != nil {
		return err
	}

	// create temp directory recursively, does not fail if directory already exists
	err = os.MkdirAll(s.TempSnapshotDir, os.ModePerm)
	if err != nil {
		return err
	}

	timestamp := time.Now().Format("_2006_01_02_15:04:05")

	// create temp file
	tmpName := "skhron" + timestamp + ".json"
	tmpFilepath := path.Join(s.TempSnapshotDir, tmpName)

	f, err := os.Create(tmpFilepath)
	if err != nil {
		return err
	}
	defer f.Close()

	// write json
	_, err = f.Write(bytes)
	if err != nil {
		return err
	}

	// create directory for snapshot file
	err = os.MkdirAll(s.SnapshotDir, os.ModePerm)
	if err != nil {
		return err
	}

	filepath := path.Join(s.SnapshotDir, s.SnapshotName+SkhronExtension)

	// if previous snapshot exists, rename it
	_, err = os.Stat(filepath)
	if err == nil {
		newPath := path.Join(s.SnapshotDir, s.SnapshotName+timestamp+SkhronExtension)
		if err := os.Rename(filepath, newPath); err != nil {
			log.Printf("failed to move old snapshot (%s) to %s\n", filepath, newPath)
		}
	}

	// move tmp file into new location
	err = os.Rename(tmpFilepath, filepath)
	if err != nil {
		return err
	}

	return nil
}

// LoadSnapshot is a function, which loads data
// from the latest snapshot file and writes data to the Skhron object.
// It looks for file {snapshot dir}/{snapshot file}.skh
// If load is failed, error is returned.
func (s *Skhron[V]) LoadSnapshot() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := path.Join(s.SnapshotDir, s.SnapshotName+SkhronExtension)

	f, err := os.Open(path)
	if err != nil {
		return err
	}

	rs := &struct {
		Data map[string]V
		TTLq *expireQueue
	}{}

	dec := json.NewDecoder(f)
	if err := dec.Decode(rs); err != nil {
		return err
	}

	// reset old skhron data
	limit := s.Data.GetLimit()
	s.Data = smap.New[string, V](limit)
	s.TTLq = newExpQueue()
	heap.Init(s.TTLq)

	for rs.TTLq.Len() > 0 {
		item := rs.TTLq.Pop().(*expireItem)
		s.Data.Set(item.Key, rs.Data[item.Key])
		s.TTLq.Push(item)
	}

	return nil
}
