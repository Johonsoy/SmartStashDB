package storage

import "sync"

type TinyWAL struct {
	option        Options
	mutex         sync.RWMutex
	activeSegment *segmentFile
}

func (w *TinyWAL) close() error {
	return nil
}
