package storage

import (
	_const "SmartStashDB/const"
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

func (w *TinyWAL) PendingWrites(data []byte) error {
	w.pendingWritesLock.Lock()
	defer w.pendingWritesLock.Unlock()

	w.maxWriteSize(int64(len(data)))
	return nil
}

func (w *TinyWAL) WriteAll() error {
	return nil
}

func (w *TinyWAL) Sync() error {
	return nil
}

func (w *TinyWAL) maxWriteSize(size int64) {
	//TODO
	return int64(_const.ChunkHeadSize + _const.Delta + (_const.Delta/_const.BlockSize+1)*_const.ChunkHeadSize)
}
