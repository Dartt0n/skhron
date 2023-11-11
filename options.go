package skhron

var (
	SkhronExtension = ".skh"
)

type StorageOpt[V any] func(s *Skhron[V])

func DefaultOpts[V any](s *Skhron[V]) {
	s.SnapshotDir = ".skhron"
	s.SnapshotName = "snapshot"
	s.TempSnapshotDir = "/tmp/skhron"

	s.Data.SetLimit(10000) // shrink map after every 10k deletions
}

func WithSnapshotDir[V any](dir string) StorageOpt[V] {
	return func(s *Skhron[V]) {
		s.SnapshotDir = dir
	}
}

func WithSnapshotName[V any](name string) StorageOpt[V] {
	return func(s *Skhron[V]) {
		s.SnapshotName = name
	}
}

func WithTempSnapshotDir[V any](dir string) StorageOpt[V] {
	return func(s *Skhron[V]) {
		s.TempSnapshotDir = dir
	}
}

func WithMapLimit[V any](limit uint64) StorageOpt[V] {
	return func(s *Skhron[V]) {
		s.Data.SetLimit(limit)
	}
}
