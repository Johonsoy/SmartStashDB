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
	options  []*Options
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

func tempDBDir() string {
	temp, _ := os.MkdirTemp("", "db-temp")
	return temp
}

type WriteOptions struct {
	Sync       bool
	DisableWal bool
}
