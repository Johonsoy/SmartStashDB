package storage

import (
	_const "SmartStashDB/const"
	"fmt"
	lru "github.com/hashicorp/golang-lru/v2"
	"os"
	"path/filepath"
)

type SegmentFileId = uint32

type SegmentFile struct {
	segmentFileId SegmentFileId

	fd *os.File

	lastBlockIndex uint32

	lastBlockSize uint32

	header []byte

	closed bool

	localCache *lru.Cache[uint32, []byte]
}

func (f *SegmentFile) readInternal(index uint32, offset uint32) ([]byte, *ChunkPosition, error) {

	return nil, nil, nil
}

func (f *SegmentFile) NewSegmentReader() *SegmentReader {
	return &SegmentReader{
		seg:         f,
		blockidx:    0,
		chunkoffset: 0,
	}
}

func (f *SegmentFile) Close() error {
	if f.closed {
		return nil
	}
	f.closed = true
	return f.fd.Close()
}

func (f *SegmentFile) Size() int64 {
	return int64(f.lastBlockIndex*_const.BlockSize + f.lastBlockSize)
}

func (f *SegmentFile) Sync() error {
	return f.fd.Sync()
}

func (f *SegmentFile) Write(data []byte) (*ChunkPosition, error) {
	return nil, nil
}

func (f *SegmentFile) WriteAll(writes [][]byte) ([]*ChunkPosition, error) {
	return nil, nil
}

func segmentFileName(dir, ext string, id uint32) string {
	return filepath.Join(dir, fmt.Sprintf("%010d"+ext, id))
}

func openSegmentFile(dir string, ext string, id uint32, localCache *lru.Cache[uint32, []byte]) (*SegmentFile, error) {
	fd, err := os.OpenFile(segmentFileName(dir, ext, id), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return nil, err
	}
	stat, err := fd.Stat()

	if err != nil {
		err := fd.Close()
		if err != nil {
			return nil, err
		}
		return nil, err
	}
	size := stat.Size()
	return &SegmentFile{
		segmentFileId:  id,
		fd:             fd,
		lastBlockIndex: uint32(size / _const.BlockSize),
		lastBlockSize:  uint32(size % _const.BlockSize),
		header:         make([]byte, _const.ChunkHeadSize),
		localCache:     localCache,
	}, nil
}
