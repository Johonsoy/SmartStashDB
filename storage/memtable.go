package storage

import (
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/dgraph-io/badger/skl"
	"github.com/dgraph-io/badger/y"
	"math"
	"os"
	"sort"
	"sync"
)

const (
	initTableId = 1

	walFileExt = ".MEM.%d"
)

type MemTable struct {
	option memTableOptions

	mu sync.RWMutex

	skl *skl.Skiplist

	tinyWal *TinyWAL
}

type memTableOptions struct {
	sklMemSize      uint32 // skip-list memory size.
	id              int    // skip-list memory id.
	walDir          string // file dir.
	walCacheSize    uint32 // wal cache size.
	walIsSync       bool   // whether to flush the disk immediately.
	walBytesPerSync uint32 // how bytes to flush the disk.
}

func openAllMemTables(options WalOptions) ([]*MemTable, error) {
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

	tables := make([]*MemTable, len(tableIds))

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

	return tables, nil
}

func openMemTable(option memTableOptions) (*MemTable, error) {
	skipList := skl.NewSkiplist(int64(option.sklMemSize * 2))

	table := &MemTable{
		option: option,
		skl:    skipList,
	}
	//TODO read wal and fill skip list
	wal, err := OpenTinyWAL(WalOptions{
		DirPath:        option.walDir,
		MemTableSize:   math.MaxInt32,
		segmentFileExt: fmt.Sprintf(walFileExt, option.id),
		Sync:           option.walIsSync,
		BytesPerSync:   option.walBytesPerSync,
		BlockCache:     option.walCacheSize,
	})
	if err != nil {
		return nil, err
	}
	table.tinyWal = wal

	indexRecords := make(map[uint64][]*LogRecord)

	reader := wal.NewReader()

	for {
		reader.Next()
	}

	return table, nil
}

func (mt *MemTable) get(key []byte) (bool, []byte) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	valueStruct := mt.skl.Get(y.KeyWithTs(key, 0))
	deleted := valueStruct.Meta == LogRecordDeleted

	return deleted, valueStruct.Value
}

func (mt *MemTable) isFull() bool {
	return mt.skl.MemSize() >= int64(mt.option.sklMemSize)
}

func (mt *MemTable) putBatch(records map[string]*LogRecord, batchId snowflake.ID, options *WriteOptions) error {
	if options == nil || options.DisableWal {
		for _, record := range records {
			record.BatchId = uint64(batchId)
			if err := mt.tinyWal.PendingWrites(record.Encode()); err != nil {
				return err
			}
		}
		record := NewLogRecord()
		record.Key = batchId.Bytes()
		record.Type = LogRecordBatchEnd

		if err := mt.tinyWal.PendingWrites(record.Encode()); err != nil {
			return err
		}

		if err := mt.tinyWal.WriteAll(); err != nil {
			return err
		}

		if options != nil && options.Sync && mt.option.walIsSync {
			if err := mt.tinyWal.Sync(); err != nil {
				return err
			}
		}
	}

	mt.mu.Lock()
	for key, record := range records {
		mt.skl.Put(y.KeyWithTs([]byte(key), 0),
			y.ValueStruct{
				Meta:  record.Type,
				Value: record.Value,
			})
	}
	mt.mu.Unlock()
	return nil
}

func (mt *MemTable) close() error {
	if mt.skl != nil {
		return mt.tinyWal.close()
	}
	return nil
}
