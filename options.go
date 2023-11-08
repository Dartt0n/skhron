package skhron

var (
	SkhronExtension = ".skh"
)

type StorageOpt func(s *Skhron)

func DefaultOpts(s *Skhron) {
	s.SnapshotDir = ".skhron"
	s.SnapshotName = "snapshot"
	s.TempSnapshotDir = "/tmp/skhron"
}

func WithSnapshotDir(dir string) StorageOpt {
	return func(s *Skhron) {
		s.SnapshotDir = dir
	}
}

func WithSnapshotName(name string) StorageOpt {
	return func(s *Skhron) {
		s.SnapshotName = name
	}
}

func WithTempSnapshotDir(dir string) StorageOpt {
	return func(s *Skhron) {
		s.TempSnapshotDir = dir
	}
}

func WithMapLimit(limit uint64) StorageOpt {
	return func(s *Skhron) {
		s.Data.SetLimit(limit)
	}
}
