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

func (w *TinyWAL) maxWriteSize(size int64) int64 {
	chunks := (size + _const.BlockSize - 1) / _const.BlockSize // 计算正确的块数（向上取整）
	total := chunks * _const.ChunkHeadSize                     // 总块头大小
	newHeadSize := _const.ChunkHeadSize + size                 // 基础头+数据大小
	return newHeadSize + total
}
