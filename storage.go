package main

import (
	"sync"

	"errors"
)

type Storage struct {
	data    map[string][]byte
	_datamu sync.Mutex
}

// NewStorage function returns a new instance of the Storage struct
// with an initialized data map and a mutex.
func NewStorage() *Storage {
	return &Storage{
		data:    make(map[string][]byte),
		_datamu: sync.Mutex{},
	}
}

// Put is a function which puts a value in the data map under a key.
// It takes the key as string and the value as byte slice.
// This function locks _datamu mutex for its operations.
func (s *Storage) Put(key string, value []byte) error {
	s._datamu.Lock()
	defer s._datamu.Unlock()

	s.data[key] = value

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
