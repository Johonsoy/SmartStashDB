package storage

import (
	_const "SmartStashDB/const"
	"io"
)

type SegmentReader struct {
	seg         *SegmentFile
	blockidx    uint32
	chunkoffset uint32
}

func (s *SegmentReader) Next() ([]byte, *ChunkPosition, error) {
	if s.seg.closed {
		return nil, nil, io.EOF
	}

	curChunk := &ChunkPosition{
		SegmentFileId: s.seg.segmentFileId,
		BlockIndex:    s.blockidx,
		ChunkOffset:   s.chunkoffset,
	}
	data, nextChunk, err := s.seg.readInternal(curChunk.BlockIndex, curChunk.ChunkOffset)
	if err != nil {
		return nil, nil, err
	}
	curChunk.ChunkSize = nextChunk.BlockIndex*_const.BlockSize + nextChunk.ChunkOffset -
		(s.chunkoffset*_const.BlockSize + curChunk.ChunkOffset)
	s.blockidx = nextChunk.BlockIndex
	s.chunkoffset = curChunk.ChunkOffset
	return data, curChunk, nil
}
