package storage

type Options struct {
	DirPath      string
	MemTableSize uint32
	Sync         bool
	BytesPerSync uint32
}
