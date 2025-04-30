package storage

import (
	_const "SmartStashDB/const"
	"errors"
	"github.com/gofrs/flock"
	"os"
	"path/filepath"
	"sync"
)

const (
	FileLockName = "FLOCK"
)

type DB struct {
	m            sync.RWMutex
	activeMem    *MemTable   // Active memory
	immutableMem []*MemTable // Immutable memory
	Closed       bool

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
	db.Closed = true
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

func (db *DB) Get(key string) ([]byte, error) {
	batch := db.batchPool.Get().(*Batch)
	batch.init(true, false, db)
	defer func() {
		_ = batch.commit(nil)
		batch.reset()
		db.batchPool.Put(batch)
	}()
	return batch.Get([]byte(key))
}

func (db *DB) getMemTables() []*MemTable {
	return db.immutableMem
}

func (db *DB) Delete(key []byte) error {
	if len(key) == 0 {
		return _const.ErrorKeyIsEmpty
	}
	batch := db.batchPool.Get().(*Batch)
	batch.init(false, false, db)
	defer func() {
		batch.reset()
		db.batchPool.Put(batch)
	}()
	return batch.delete(key)
}

func OpenDB(options Options) (*DB, error) {

	// Check if file existed.
	if _, err := os.Stat(options.DirPath); err != nil {
		if err := os.Mkdir(options.DirPath, os.ModePerm); err != nil {
			return nil, err
		}
	}
	lock, err := flock.New(filepath.Join(options.DirPath, FileLockName)).TryLock()
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
