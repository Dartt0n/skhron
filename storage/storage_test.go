package storage

import (
	"context"
	"encoding/binary"
	"testing"
	"time"
)

func TestStoragePutGetNew(t *testing.T) {
	t.Parallel()
	storage := New()

	var testValue uint32 = 150645

	bytes := make([]byte, 4)
	binary.NativeEndian.PutUint32(bytes, testValue)

	if err := storage.Put("test-new", bytes, time.Second); err != nil {
		t.Errorf("put failed: %v", err)
	}

	bytes, err := storage.Get("test-new")
	if err != nil {
		t.Errorf("get failed: %v", err)
	}
	value := binary.NativeEndian.Uint32(bytes)

	if value != testValue {
		t.Errorf("value mismatch. expected: %d, got: %d", testValue, value)
	}
}

func TestStoragePutGetOverride(t *testing.T) {
	t.Parallel()
	storage := New()

	var testValue uint32 = 12312

	bytes := make([]byte, 4)
	binary.NativeEndian.PutUint32(bytes, 100500)

	if err := storage.Put("test-override", bytes, time.Second); err != nil {
		t.Errorf("put failed: %v", err)
	}

	binary.NativeEndian.PutUint32(bytes, testValue)
	if err := storage.Put("test-override", bytes, time.Second); err != nil {
		t.Errorf("put failed: %v", err)
	}

	bytes, err := storage.Get("test-override")
	if err != nil {
		t.Errorf("get failed: %v", err)
	}
	value := binary.NativeEndian.Uint32(bytes)

	if value != testValue {
		t.Errorf("value mismatch. expected: %d, got: %d", testValue, value)
	}
}

func TestStorageGetUnknown(t *testing.T) {
	t.Parallel()
	s := New()

	value, err := s.Get("test-get-unknown")
	if err == nil {
		t.Errorf("get unknown failed. got: %v", value)
	}
}

func TestStorageDeleteExisting(t *testing.T) {
	t.Parallel()
	storage := New()

	var testValue uint32 = 150645

	bytes := make([]byte, 4)
	binary.NativeEndian.PutUint32(bytes, testValue)

	if err := storage.Put("test-delete-existing", bytes, time.Second); err != nil {
		t.Errorf("put failed: %v", err)
	}

	err := storage.Delete("test-delete-existing")
	if err != nil {
		t.Errorf("delete failed: %v", err)
	}

	_, err = storage.Get("test-delete-existing")
	if err == nil {
		t.Errorf("delete failed. expected key \"test-delete-existing\" not to exist: %v", err)
	}
}

func TestStorageDeleteNonExisting(t *testing.T) {
	t.Parallel()
	s := New()

	err := s.Delete("test-delete-nonexisting")
	if err != nil {
		t.Errorf("delete failed: %v", err)
	}
}

func TestStorageExistExisting(t *testing.T) {
	t.Parallel()
	storage := New()

	var testValue uint32 = 231

	bytes := make([]byte, 4)
	binary.NativeEndian.PutUint32(bytes, testValue)

	if err := storage.Put("test-exist-existing", bytes, time.Second); err != nil {
		t.Errorf("put failed: %v", err)
	}

	value := storage.Exists("test-exist-existing")
	if value != true {
		t.Errorf("exists failed")
	}
}

func TestStorageExistNonExisting(t *testing.T) {
	t.Parallel()
	s := New()

	value := s.Exists("test-exist-non-existing")
	if value != false {
		t.Errorf("exists failed")
	}
}

func TestStorageCleanup(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	storage := New()
	go storage.CleaningProcess(ctx, 500*time.Millisecond, done)

	if err := storage.Put("test", []byte("hello world"), 500*time.Millisecond); err != nil {
		t.Errorf("put failed: %v", err)
	}

	if !storage.Exists("test") {
		t.Errorf("exists failed")
	}

	time.Sleep(1 * time.Second)
	cancel()
	<-done

	// clean up should be performed by this time
	if storage.Exists("test") {
		t.Errorf("cleanup failed")
	}
}
