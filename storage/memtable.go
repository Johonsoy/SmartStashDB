package storage

import (
	"fmt"
	"os"
	"sort"
)

const (
	initTableId = 1

	walFileExt = ".MEM.%d"
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
	dir, err := os.ReadDir(options.DirPath)
	if err != nil {
		return nil, err
	}
	var tableIds []int

	for _, file := range dir {
		if file.IsDir() {
			continue
		}
		var id int
		var prefix int

		_, err = fmt.Sscanf(file.Name(), "memtable_%d"+walFileExt, &prefix, &id)
		if err != nil {
			continue
		}
		tableIds = append(tableIds, id)
	}

	if len(tableIds) == 0 {
		tableIds = append(tableIds, initTableId)
	}

	sort.Ints(tableIds)

	tables := make([]*memTable, len(tableIds))

	for i, id := range tableIds {
		table, err := openMemTable()
	}

	return nil, nil
}

func openMemTable() (*memTable, error) {
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
