package _const

import (
	"SmartStashDB/storage"
	"io"
)

type Reader struct {
	AllSegmentReader []*storage.SegmentReader
	Progress         int
}

func (r *Reader) Next() ([]byte, *storage.ChunkPosition, error) {
	if r.Progress >= len(r.AllSegmentReader) {
		return nil, nil, io.EOF
	}
	data, chunkPos, err := r.AllSegmentReader[r.Progress].Next()
	if err == io.EOF {
		r.Progress++
		return r.Next()
	}
	return data, chunkPos, err
}
