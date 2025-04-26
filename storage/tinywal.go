package storage


type TinyWAL struct {
	option        Options
	mutex         sync.RWMutex
	activeSegment *tinywal.segmentFile
}

func (w *TinyWAL) close() error {
	return nil
}
