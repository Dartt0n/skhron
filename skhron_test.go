package skhron

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestStorage_PutGetNew(t *testing.T) {
	t.Parallel()
	storage := New[int]()

	testValue := 150645

	if err := storage.Put("test-new", testValue, time.Second); err != nil {
		t.Errorf("put failed: %v", err)
	}

	value, err := storage.Get("test-new")
	if err != nil {
		t.Errorf("get failed: %v", err)
	}

	if value != testValue {
		t.Errorf("value mismatch. expected: %d, got: %d", testValue, value)
	}
}

func TestStorage_PutGetOverride(t *testing.T) {
	t.Parallel()
	storage := New[int]()

	testValue := 12312

	if err := storage.Put("test-override", testValue, time.Second); err != nil {
		t.Errorf("put failed: %v", err)
	}

	if err := storage.Put("test-override", testValue, time.Second); err != nil {
		t.Errorf("put failed: %v", err)
	}

	value, err := storage.Get("test-override")
	if err != nil {
		t.Errorf("get failed: %v", err)
	}

	if value != testValue {
		t.Errorf("value mismatch. expected: %d, got: %d", testValue, value)
	}
}

func TestStorage_GetUnknown(t *testing.T) {
	t.Parallel()
	s := New[string]()

	value, err := s.Get("test-get-unknown")
	if err == nil {
		t.Errorf("get unknown failed. got: %v", value)
	}
}

func TestStorage_DeleteExisting(t *testing.T) {
	t.Parallel()
	storage := New[int]()

	testValue := 150645

	if err := storage.Put("test-delete-existing", testValue, time.Second); err != nil {
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

func TestStorage_DeleteNonExisting(t *testing.T) {
	t.Parallel()
	s := New[int]()

	err := s.Delete("test-delete-nonexisting")
	if err != nil {
		t.Errorf("delete failed: %v", err)
	}
}

func TestStorage_ExistExisting(t *testing.T) {
	t.Parallel()
	storage := New[int]()

	testValue := 231

	if err := storage.Put("test-exist-existing", testValue, time.Second); err != nil {
		t.Errorf("put failed: %v", err)
	}

	value := storage.Exists("test-exist-existing")
	if value != true {
		t.Errorf("exists failed")
	}
}

func TestStorage_ExistNonExisting(t *testing.T) {
	t.Parallel()
	s := New[int]()

	value := s.Exists("test-exist-non-existing")
	if value != false {
		t.Errorf("exists failed")
	}
}

func TestStorage_Cleanup(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	storage := New[string]()
	go storage.PeriodicCleanup(ctx, 500*time.Millisecond, done)

	if err := storage.Put("test", "hello world", 500*time.Millisecond); err != nil {
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

func TestStorage_StructGeneric(t *testing.T) {
	type TestStruct struct {
		msg string
		age int
	}

	s := New[TestStruct]()

	testValue := TestStruct{
		msg: "hello world",
		age: 5,
	}

	s.Put("test-value", testValue, 1*time.Hour)

	if v, err := s.Get("test-value"); err != nil || v != testValue {
		t.Errorf("get failed: %v", err)
	}
}

