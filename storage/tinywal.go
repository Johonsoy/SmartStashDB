package storage

import (
	"SmartStashDB/tinywal"
	lru "github.com/hashicorp/golang-lru/v2"
	"sync"
)

type TinyWAL struct {
	option           Options
	mutex            sync.RWMutex
	activeSegment    *tinywal.SegmentFile
	immutableSegment map[SegmentFileId]*tinywal.SegmentFile
	localCache       *lru.Cache[uint64, []byte]
	byteWrite        uint64

	pendingWritesLock sync.Mutex
	pendingWrites     [][]byte
	pendingWritesSize uint64
}

func (w *TinyWAL) close() error {
	return nil
}

func (w *TinyWAL) PendingWrites(encode []byte) error {
	return nil
}

func (w *TinyWAL) WriteAll() error {
	return nil
}

func (w *TinyWAL) Sync() error {
	return nil
}
