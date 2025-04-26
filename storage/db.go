package storage

import (
	"errors"
	"github.com/gofrs/flock"
	"os"
	"path/filepath"
	"sync"
)

const (
	fileLockName = "FLOCK"
)

type DB struct {
	m            sync.RWMutex
	activeMem    *memTable   // Active memory
	immutableMem []*memTable // Immutable memory
	closed       bool

	batchPool sync.Pool
}

func (db *DB) Close() error {
	db.m.Lock()
	defer db.m.Unlock()

	for _, table := range db.immutableMem {
		err := table.close()
		if err != nil {
			return err
		}
	}
	if err := db.activeMem.close(); err != nil {
		return err
	}
	db.closed = true
	return nil
}

func (db *DB) Put(key string, value string, options *WriteOptions) error {
	batch := db.batchPool.Get().(*Batch)
	defer func() {
		batch.reset()
		db.batchPool.Put(batch)
	}()
	batch.init(false, false, db).writePendingWrites()
	err := batch.put([]byte(key), []byte(value))
	if err != nil {
		batch.unLock()
		return err
	}
	return batch.commit(options)
}

func (db *DB) waitMemTableSpace() error {
	if db.activeMem.isFull() {
		return nil
	}
	db.immutableMem = append(db.immutableMem, db.activeMem)
	option := db.activeMem.option
	option.id++
	table, err := openMemTable(option)
	if err != nil {
		return err
	}
	db.activeMem = table
	return nil
}

func OpenDB(options Options) (*DB, error) {

	// Check if file existed.
	if _, err := os.Stat(options.DirPath); err != nil {
		if err := os.Mkdir(options.DirPath, os.ModePerm); err != nil {
			return nil, err
		}
	}
	lock, err := flock.New(filepath.Join(options.DirPath, fileLockName)).TryLock()
	if err != nil {
		return nil, err
	}
	if !lock {
		return nil, errors.New("file locked")
	}

	memTables, err := openAllMemTables(options)
	if err != nil {
		return nil, err
	}
	db := &DB{
		activeMem:    memTables[len(memTables)-1],
		immutableMem: memTables,
		batchPool:    sync.Pool{New: makeBatch},
	}
	return db, nil
}
