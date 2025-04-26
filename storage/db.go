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

func (db *DB) close() error {
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
