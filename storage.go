package main

import (
	"sync"

	"errors"
)

type Storage struct {
	data    map[string][]byte
	_datamu sync.Mutex
}

func NewStorage() *Storage {
	return &Storage{
		data:    make(map[string][]byte),
		_datamu: sync.Mutex{},
	}
}

func (s *Storage) Put(key string, value []byte) error {
	s._datamu.Lock()
	defer s._datamu.Unlock()

	s.data[key] = value

	return nil
}

func (s *Storage) Get(key string) ([]byte, error) {
	s._datamu.Lock()
	defer s._datamu.Unlock()

	if v, ok := s.data[key]; ok {
		return v, nil
	}

	return []byte{}, errors.New("no such key: " + key)
}

func (s *Storage) Delete(key string) error {
	s._datamu.Lock()
	defer s._datamu.Unlock()

	delete(s.data, key)

	return nil
}

func (s *Storage) Exists(key string) bool {
	s._datamu.Lock()
	defer s._datamu.Unlock()

	_, exist := s.data[key]
	return exist
}
