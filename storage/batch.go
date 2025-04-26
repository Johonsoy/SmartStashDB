package storage

import "sync"
import "github.com/bwmarrin/snowflake"

func makeBatch() interface{} {
	return nil
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

}

func (batch *Batch) writePendingWrites() *Batch {
	batch.pendingWrites = make(map[string]*LogRecord)
	return batch
}

func (batch *Batch) put(bytes []byte, bytes2 []byte) error {
	return nil
}

func (batch *Batch) unblock() {

}

func (batch *Batch) commit() error {
	return nil
}
