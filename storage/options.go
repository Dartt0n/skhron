package storage

var (
	SkhronExtension     = ".skh"
	SkronTmpSnapshotDir = "/tmp/skhron"
)

type StorageOpt func(s *Storage)

func DefaultOpts(s *Storage) {
	s.SnapshotDir = ".skhron"
	s.SnapshotName = "snapshot"
	s.TempSnapshotDir = SkronTmpSnapshotDir
}

func WithSnapshotDir(dir string) StorageOpt {
	return func(s *Storage) {
		s.SnapshotDir = dir
	}
}

func WithSnapshotName(name string) StorageOpt {
	return func(s *Storage) {
		s.SnapshotName = name
	}
}

func WithTempSnapshotDir(dir string) StorageOpt {
	return func(s *Storage) {
		s.TempSnapshotDir = dir
	}
}
