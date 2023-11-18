package skhron

import (
	"context"
	"reflect"
	"testing"
	"time"
)

// Unit tests

func TestNew(t *testing.T) {
	type args struct {
		opts []StorageOpt[int]
	}

	snapshotNameEq := func(name string) func(s *Skhron[int]) bool {
		return func(s *Skhron[int]) bool {
			return reflect.DeepEqual(s.SnapshotName, name)
		}
	}

	snapshotDirEq := func(dir string) func(s *Skhron[int]) bool {
		return func(s *Skhron[int]) bool {
			return reflect.DeepEqual(s.SnapshotDir, dir)
		}
	}

	limitEq := func(limit uint64) func(s *Skhron[int]) bool {
		return func(s *Skhron[int]) bool {
			return reflect.DeepEqual(s.Data.GetLimit(), limit)
		}
	}

	tests := []struct {
		name  string
		args  args
		check func(s *Skhron[int]) bool
	}{
		{
			"custom snapshot name",
			args{[]StorageOpt[int]{WithSnapshotName[int]("custom-snapshot-name")}},
			snapshotNameEq("custom-snapshot-name"),
		},
		{
			"empty snapshot name",
			args{[]StorageOpt[int]{WithSnapshotName[int]("")}},
			snapshotNameEq(""),
		},
		{
			"snapshot name with dot",
			args{[]StorageOpt[int]{WithSnapshotName[int]("name.backup")}},
			snapshotNameEq("name.backup"),
		},
		{
			"path in snapshot name", // todo: should fail
			args{[]StorageOpt[int]{WithSnapshotName[int]("name/backup")}},
			snapshotNameEq("name/backup"),
		},
		{
			"custom snapshot dir",
			args{[]StorageOpt[int]{WithSnapshotDir[int]("/data/skhron/backup")}},
			snapshotDirEq("/data/skhron/backup"),
		},
		{
			"local snapshot dir",
			args{[]StorageOpt[int]{WithSnapshotDir[int](".")}},
			snapshotDirEq("."),
		},
		{
			"home snapshot dir",
			args{[]StorageOpt[int]{WithSnapshotDir[int]("~")}},
			snapshotDirEq("~"),
		},
		{
			"limit 1",
			args{[]StorageOpt[int]{WithMapLimit[int](1)}},
			limitEq(1),
		},
		{
			"limit 100",
			args{[]StorageOpt[int]{WithMapLimit[int](100)}},
			limitEq(100),
		},
		{
			"limit 0", // todo: should fail
			args{[]StorageOpt[int]{WithMapLimit[int](0)}},
			limitEq(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.opts...); !tt.check(got) {
				t.Errorf("invalid struct from New(): %v", got)
			}
		})
	}
}

func TestPutTTL(t *testing.T) {
	// Create a new instance of Skhron
	s := New[string]()

	// Add some test data
	s.PutTTL("key1", "value1", time.Minute)
	s.PutTTL("key2", "value2", time.Minute)
	s.PutTTL("key3", "value3", time.Minute)

	// Test cases
	tests := []struct {
		name      string
		key       string
		value     string
		ttl       time.Duration
		queue_len int
		queue_ind int
	}{
		{
			name:      "NewKey",
			key:       "key4",
			value:     "value4",
			ttl:       time.Minute,
			queue_len: 4, // item is pushed at last position
			queue_ind: 3,
		},
		{
			name:      "ExistingKey",
			key:       "key1",
			value:     "updatedValue1",
			ttl:       time.Hour,
			queue_len: 4, // item is not pushed, but moved to end
			queue_ind: 3,
		},
		{
			name:      "EmptyKey",
			key:       "",
			value:     "value",
			ttl:       30 * time.Second,
			queue_len: 5, // item is pushed at front
			queue_ind: 0,
		},
		{
			name:      "EmptyValue",
			key:       "key5",
			value:     "",
			ttl:       45 * time.Second,
			queue_len: 6, // item is pushed in middle
			queue_ind: 2,
		},
		{
			name:      "ZeroTTL",
			key:       "key3",
			value:     "value",
			ttl:       0,
			queue_len: 6, // item is moved t the front
			queue_ind: 0,
		},
	}

	// Run the tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the Put function
			err := s.PutTTL(tt.key, tt.value, tt.ttl)

			// Check if the Put function returned an error
			if err != nil {
				t.Errorf("%s : Put() returned an error: %v", tt.name, err)
			}

			// Check if the value is updated in the map
			if s.Data.Get(tt.key) != tt.value {
				t.Errorf("%s : Value in the map after Put() = %v, want %v", tt.name, s.Data.Get(tt.key), tt.value)
			}

			// Check if the queue length matches the expected value
			if s.TTLq.Len() != tt.queue_len {
				t.Errorf("%s : Length of queue after Put() = %v, want %v", tt.name, s.TTLq.Len(), tt.queue_len)
			}

			// Check if the item is pushed to the queue when it's a new key
			if s.TTLq.Len() > 0 && (*s.TTLq)[tt.queue_ind].Key != tt.key {
				t.Errorf("%s : Item with key %v not pushed to the queue", tt.name, tt.key)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	s := New[string]()

	s.Data.Set("key1", "value1")
	s.Data.Set("key2", "value2")
	s.Data.Set("key3", "value3")

	tests := []struct {
		name        string
		keyToDelete string
	}{
		{name: "NonExistingKey", keyToDelete: "key0"},
		{name: "ExistingKey", keyToDelete: "key1"},
		{name: "ExistingKey", keyToDelete: "key2"},
		{name: "ExistingKey", keyToDelete: "key3"},
		{name: "NonExistingKey", keyToDelete: "key4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Delete(tt.keyToDelete)

			if err != nil {
				t.Errorf("Delete() returned an error: %v", err)
			}

			if _, ok := s.Data.Get2(tt.keyToDelete); ok == true {
				t.Errorf("Deleted key %v is still present in the map", tt.keyToDelete)
			}
		})
	}
}

func TestExists(t *testing.T) {
	// Create a new instance of Skhron
	s := New[string]()

	// Add some test data
	s.Data.Set("key1", "value1")
	s.Data.Set("key2", "value2")
	s.Data.Set("key3", "value3")

	// Test cases
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{name: "NonExistingKey", key: "key0", expected: false},
		{name: "ExistingKey", key: "key1", expected: true},
		{name: "ExistingKey", key: "key2", expected: true},
		{name: "ExistingKey", key: "key3", expected: true},
		{name: "NonExistingKey", key: "key4", expected: false},
	}

	// Run the tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the Exists function
			exist := s.Exists(tt.key)

			// Check if the result matches the expected value
			if exist != tt.expected {
				t.Errorf("Exists() = %v, want %v", exist, tt.expected)
			}
		})
	}
}

func TestSkhronGet(t *testing.T) {
	// Set up test data or dependencies
	s := New[int]()

	s.Data.Set("key1", 1)
	s.Data.Set("key2", 3)
	s.Data.Set("key3", 4)

	tests := []struct {
		name     string
		key      string
		expected bool
		value    int
	}{
		{name: "NonExistingKey", key: "key0", expected: false},
		{name: "ExistingKey", key: "key1", expected: true, value: 1},
		{name: "ExistingKey", key: "key2", expected: true, value: 3},
		{name: "ExistingKey", key: "key3", expected: true, value: 4},
		{name: "NonExistingKey", key: "key4", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the Exists function
			value, err := s.Get(tt.key)

			if err != nil && tt.expected {
				t.Errorf("Get(%s) = %v, wanted value", tt.key, err)
			}

			if value != tt.value {
				t.Errorf("Get(%s) = %d, wanted %d", tt.key, value, tt.value)
			}
		})
	}
}

// Scenario Tests

func TestPutGetNew(t *testing.T) {
	t.Parallel()
	storage := New[int]()

	testValue := 150645

	if err := storage.PutTTL("test-new", testValue, time.Second); err != nil {
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

func TestPutGetOverride(t *testing.T) {
	t.Parallel()
	storage := New[int]()

	testValue := 12312

	if err := storage.PutTTL("test-override", testValue, time.Second); err != nil {
		t.Errorf("put failed: %v", err)
	}

	if err := storage.PutTTL("test-override", testValue, time.Second); err != nil {
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

func TestGetUnknown(t *testing.T) {
	t.Parallel()
	s := New[string]()

	value, err := s.Get("test-get-unknown")
	if err == nil {
		t.Errorf("get unknown failed. got: %v", value)
	}
}

func TestDeleteExisting(t *testing.T) {
	t.Parallel()
	storage := New[int]()

	testValue := 150645

	if err := storage.PutTTL("test-delete-existing", testValue, time.Second); err != nil {
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

func TestDeleteNonExisting(t *testing.T) {
	t.Parallel()
	s := New[int]()

	err := s.Delete("test-delete-nonexisting")
	if err != nil {
		t.Errorf("delete failed: %v", err)
	}
}

func TestExistExisting(t *testing.T) {
	t.Parallel()
	storage := New[int]()

	testValue := 231

	if err := storage.PutTTL("test-exist-existing", testValue, time.Second); err != nil {
		t.Errorf("put failed: %v", err)
	}

	value := storage.Exists("test-exist-existing")
	if value != true {
		t.Errorf("exists failed")
	}
}

func TestExistNonExisting(t *testing.T) {
	t.Parallel()
	s := New[int]()

	value := s.Exists("test-exist-non-existing")
	if value != false {
		t.Errorf("exists failed")
	}
}

func TestCleanup(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	storage := New[string]()
	go storage.PeriodicCleanup(ctx, 500*time.Millisecond, done)

	if err := storage.PutTTL("test", "hello world", 500*time.Millisecond); err != nil {
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

func TestStructGeneric(t *testing.T) {
	type TestStruct struct {
		Msg string `json:"msg"`
		Age int    `json:"age"`
	}

	s := New[TestStruct]()

	testValue := TestStruct{
		Msg: "hello world",
		Age: 5,
	}

	s.PutTTL("test-value", testValue, 1*time.Hour)

	if v, err := s.Get("test-value"); err != nil || v != testValue {
		t.Errorf("get failed: %v", err)
	}

	if err := s.CreateSnapshot(); err != nil {
		t.Errorf("create snapshot failed: %v", err)
	}
}
