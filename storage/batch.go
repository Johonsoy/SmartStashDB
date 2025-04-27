package storage

import (
	_const "SmartStashDB/const"
	"sync"
)
import "github.com/bwmarrin/snowflake"

func makeBatch() interface{} {
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	return &Batch{
		options: DefaultBatchOptions,
		m:       sync.RWMutex{},
		batchId: node,
	}
}

type Batch struct {
	db            *DB
	pendingWrites map[string]*LogRecord
	options       BatchOptions
	m             sync.RWMutex
	commited      bool
	batchId       *snowflake.Node
}

func (batch *Batch) reset() {

}

func (batch *Batch) init(readOnly bool, sync bool, db *DB) *Batch {
	batch.db = db
	batch.options.ReadOnly = readOnly
	batch.options.Sync = sync
	batch.lock()
	return batch
}

func (batch *Batch) lock() {
	if batch.options.ReadOnly {
		batch.db.m.RLock()
	} else {
		batch.db.m.Lock()
	}
}

func (batch *Batch) writePendingWrites() *Batch {
	batch.pendingWrites = make(map[string]*LogRecord)
	return batch
}

func (batch *Batch) put(key []byte, value []byte) error {
	if len(key) == 0 {
		return _const.ErrorKeyIsEmpty
	}

	if batch.db.Closed {
		return _const.ErrorDBClosed
	}

	if batch.options.ReadOnly {
		return _const.ErrorReadOnlyBatch
	}
	batch.m.Lock()
	defer batch.m.Unlock()
	batch.pendingWrites[string(key)] = &LogRecord{
		Key:   key,
		Value: value,
		Type:  LogRecordNormal,
	}
	return nil
}

func (batch *Batch) unLock() {
	if batch.options.ReadOnly {
		batch.db.m.RUnlock()
	} else {
		batch.db.m.Unlock()
	}

}

func (batch *Batch) commit(w *WriteOptions) error {
	if w == nil {
		w = &WriteOptions{
			Sync:       false,
			DisableWal: false,
		}
	}
	defer batch.unLock()
	if batch.db.Closed {
		return _const.ErrorDBClosed
	}

	if batch.options.ReadOnly || len(batch.pendingWrites) == 0 {
		return nil
	}

	batch.m.Lock()
	defer batch.m.Unlock()
	if batch.commited {
		return _const.ErrorBatchCommited
	}

	if err := batch.db.waitMemTableSpace(); err != nil {
		return err
	}

	batchId := batch.batchId.Generate()
	if err := batch.db.activeMem.putBatch(batch.pendingWrites, batchId, w); err != nil {
		return err
	}
	batch.commited = true
	return nil
}
