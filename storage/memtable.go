package storage

const (
	initTableId = 1
)

type memTable struct {
}

type memTableOptions struct {
	sklMemSize      uint32 // skip-list memory size.
	id              int    // skip-list memory id.
	walDir          string // file dir.
	walCacheSize    int    // wal cache size.
	walIsSync       bool   // whether to flush the disk immediately.
	walBytesPerSync uint64 // how bytes to flush the disk.
}

func openAllMemTables(options Options) ([]*memTable, error) {
	return nil, nil
}

func (mt *memTable) get(key []byte) ([]byte, error) {
	return nil, nil
}

func (mt *memTable) putBatch() error {
	return nil
}

func (mt *memTable) close() error {
	return nil
}
