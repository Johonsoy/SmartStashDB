package storage

import (
	"fmt"
	"github.com/dgraph-io/badger/skl"
	"github.com/dgraph-io/badger/y"
	"os"
	"sort"
	"sync"
)

const (
	initTableId = 1

	walFileExt = ".MEM.%d"
)

type memTable struct {
	option memTableOptions

	mu sync.RWMutex

	skl *skl.Skiplist

	tinyWal *TinyWAL
}

type memTableOptions struct {
	sklMemSize      uint32 // skip-list memory size.
	id              int    // skip-list memory id.
	walDir          string // file dir.
	walCacheSize    int    // wal cache size.
	walIsSync       bool   // whether to flush the disk immediately.
	walBytesPerSync uint32 // how bytes to flush the disk.
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
		table, err := openMemTable(memTableOptions{
			sklMemSize:      options.MemTableSize,
			id:              id,
			walDir:          options.DirPath,
			walIsSync:       options.Sync,
			walBytesPerSync: options.BytesPerSync,
		})

		if err != nil {
			return nil, err
		}
		tables[i] = table
	}

	return nil, nil
}

func openMemTable(options memTableOptions) (*memTable, error) {
	skipList := skl.NewSkiplist(int64(options.sklMemSize * 2))

	table := &memTable{
		option: options,
		skl:    skipList,
	}
	//TODO read wal and fill skip list
	return table, nil
}

func (mt *memTable) get(key []byte) (bool, []byte) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	valueStruct := mt.skl.Get(y.KeyWithTs(key, 0))
	deleted := valueStruct.Meta == LogRecordDeleted

	return deleted, valueStruct.Value
}

func (mt *memTable) isFull() bool {
	return mt.skl.MemSize() >= int64(mt.option.sklMemSize)
}

func (mt *memTable) putBatch() error {
	return nil
}

func (mt *memTable) close() error {
	if mt.skl != nil {
		return mt.tinyWal.close()
	}
	return nil
}
