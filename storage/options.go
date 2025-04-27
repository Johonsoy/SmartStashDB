package storage

import (
	_const "SmartStashDB/const"
	"os"
)

type Options struct {
	DirPath      string
	MemTableSize uint32
	Sync         bool
	BytesPerSync uint32
	BlockCache   uint32
}

type BatchOptions struct {
	ReadOnly bool
	Sync     bool
}

var DefaultOptions = Options{
	DirPath:      tempDBDir(),
	MemTableSize: 64 * _const.MB,
	BlockCache:   0,
	Sync:         false,
	BytesPerSync: 0,
}

var DefaultBatchOptions = BatchOptions{
	ReadOnly: false,
	Sync:     true,
}

func tempDBDir() string {
	temp, _ := os.MkdirTemp("", "db-temp")
	return temp
}

type WriteOptions struct {
	Sync       bool
	DisableWal bool
}
