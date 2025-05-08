package _const

import "SmartStashDB/storage"

type Reader struct {
	allSegmentReader []*storage.SegmentReader
	progress         int
}

func (r *Reader) Next() ([]byte, *ChunkPosition, error) {
	return nil, nil, nil
}
