package storage

import (
	"SmartStashDB/tinywal"
	"sync"
)

type TinyWAL struct {
	option        Options
	mutex         sync.RWMutex
	activeSegment *tinywal.SegmentFile
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
