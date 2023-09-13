package main

import (
	"encoding/binary"
	"testing"
)

func TestStoragePutGetNew(t *testing.T) {
	s := NewStorage()

	var testValue uint32 = 150645

	bytes := make([]byte, 4)
	binary.NativeEndian.PutUint32(bytes, testValue)

	if err := s.Put("test-new", bytes); err != nil {
		t.Errorf("put failed: %v", err)
	}

	bytes, err := s.Get("test-new")
	if err != nil {
		t.Errorf("get failed: %v", err)
	}
	value := binary.NativeEndian.Uint32(bytes)

	if value != testValue {
		t.Errorf("value mismatch. expected: %d, got: %d", testValue, value)
	}
}

func TestStoragePutGetOverride(t *testing.T) {
	s := NewStorage()

	var testValue uint32 = 12312

	bytes := make([]byte, 4)
	binary.NativeEndian.PutUint32(bytes, 100500)

	if err := s.Put("test-override", bytes); err != nil {
		t.Errorf("put failed: %v", err)
	}

	binary.NativeEndian.PutUint32(bytes, testValue)
	if err := s.Put("test-override", bytes); err != nil {
		t.Errorf("put failed: %v", err)
	}

	bytes, err := s.Get("test-override")
	if err != nil {
		t.Errorf("get failed: %v", err)
	}
	value := binary.NativeEndian.Uint32(bytes)

	if value != testValue {
		t.Errorf("value mismatch. expected: %d, got: %d", testValue, value)
	}
}

func TestStorageGetUnknown(t *testing.T) {
	s := NewStorage()

	value, err := s.Get("test-get-unknown")
	if err == nil {
		t.Errorf("get unknown failed. got: %v", value)
	}
}

func TestStorageDeleteExisting(t *testing.T) {
	s := NewStorage()

	var testValue uint32 = 150645

	bytes := make([]byte, 4)
	binary.NativeEndian.PutUint32(bytes, testValue)

	if err := s.Put("test-delete-existing", bytes); err != nil {
		t.Errorf("put failed: %v", err)
	}

	err := s.Delete("test-delete-existing")
	if err != nil {
		t.Errorf("delete failed: %v", err)
	}

	_, err = s.Get("test-delete-existing")
	if err == nil {
		t.Errorf("delete failed. expected key \"test-delete-existing\" not to exist: %v", err)
	}
}

func TestStorageDeleteNonExisting(t *testing.T) {
	s := NewStorage()

	err := s.Delete("test-delete-nonexisting")
	if err != nil {
		t.Errorf("delete failed: %v", err)
	}
}
