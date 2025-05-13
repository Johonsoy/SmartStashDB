package _const

import (
	"SmartStashDB/storage"
	"io"
)

type Reader struct {
	allSegmentReader []*storage.SegmentReader
	progress         int
}

func (r *Reader) Next() ([]byte, *storage.ChunkPosition, error) {
	if r.progress >= len(r.allSegmentReader) {
		return nil, nil, io.EOF
	}
	data, chunkPos, err := r.allSegmentReader[r.progress].Next()
	if err == io.EOF {
		r.progress++
		return r.Next()
	}
	return data, chunkPos, err
}
