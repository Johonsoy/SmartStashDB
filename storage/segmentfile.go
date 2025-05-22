package storage

import (
	_const "SmartStashDB/const"
	"bytes"
	"encoding/binary"
	"fmt"
	lru "github.com/hashicorp/golang-lru/v2"
	"hash/crc32"
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
	if f.closed {
		return nil, _const.ErrClosed
	}
	index := f.lastBlockIndex
	size := f.lastBlockSize

	var err error
	buffer := DefaultBuffer.Get()
	defer func() {
		DefaultBuffer.Put(buffer)
	}()
	writeBuffer, err := f.writeBuffer(data, buffer)
	if err != nil {
		f.lastBlockIndex = index
		f.lastBlockSize = size
		return nil, err
	}
	err = f.writeBuffer2File(buffer)
	if err != nil {
		f.lastBlockIndex = index
		f.lastBlockSize = size
		return nil, err
	}
	return writeBuffer, nil
}

func (f *SegmentFile) WriteAll(writes [][]byte) (position []*ChunkPosition, err error) {
	if f.closed {
		return nil, _const.ErrClosed
	}

	index := f.lastBlockIndex
	lastBlockSize := f.lastBlockSize

	buffer := DefaultBuffer.Get()

	defer func() {
		if err != nil {
			f.lastBlockIndex = index
			f.lastBlockSize = lastBlockSize
		}
		DefaultBuffer.Put(buffer)
	}()

	positions := make([]*ChunkPosition, len(writes))
	for i, data := range writes {
		pos, err := f.writeBuffer(data, buffer)
		if err != nil {
			return nil, err
		}
		positions[i] = pos
	}

	if err := f.writeBuffer2File(buffer); err != nil {
		return nil, err
	}
	return positions, nil
}

func (f *SegmentFile) writeBuffer(bytes []byte, buffer *bytes.Buffer) (*ChunkPosition, error) {
	if f.closed {
		return nil, _const.ErrClosed
	}

	padding := uint32(0)

	// Pre-grow the buffer for better performance
	totalWriteSize := len(bytes) + int(_const.ChunkHeadSize)*2 // Estimate needed size
	buffer.Grow(totalWriteSize)

	if f.lastBlockSize+_const.ChunkHeadSize >= _const.BlockSize {
		size := _const.BlockSize - f.lastBlockSize
		_, err := buffer.Write(make([]byte, size))
		if err != nil {
			return nil, err
		}
		padding += size
		f.lastBlockIndex++
		f.lastBlockSize = 0
	}

	position := &ChunkPosition{
		SegmentFileId: f.segmentFileId,
		BlockIndex:    f.lastBlockIndex,
		ChunkOffset:   f.lastBlockSize,
	}

	dataLen := uint32(len(bytes))

	if f.lastBlockSize+_const.ChunkHeadSize <= _const.BlockSize {
		err := f.appendChunk2Buffer(buffer, bytes, ChunkTypeFull)
		if err != nil {
			return nil, err
		}
		position.ChunkSize = dataLen + _const.ChunkHeadSize
	} else {
		// Split data into many chunks across blocks
		var (
			remainingDataSize        = dataLen
			curBlockSize             = f.lastBlockSize
			chunkNum          uint32 = 0
		)

		for remainingDataSize > 0 {
			chunkType := ChunkTypeMiddle
			if remainingDataSize == dataLen {
				chunkType = ChunkTypeStart
			}
			freeSize := _const.BlockSize - curBlockSize - _const.ChunkHeadSize
			if freeSize >= remainingDataSize {
				freeSize = remainingDataSize
				chunkType = ChunkTypeEnd
			}
			err := f.appendChunk2Buffer(buffer, bytes[dataLen-remainingDataSize:dataLen-remainingDataSize+freeSize], chunkType)
			if err != nil {
				return nil, err
			}
			chunkNum++
			remainingDataSize -= freeSize
			curBlockSize = (curBlockSize + _const.ChunkHeadSize + freeSize) % _const.BlockSize
		}
		position.ChunkSize = chunkNum*_const.ChunkHeadSize + dataLen
	}

	return position, nil
}

func (f *SegmentFile) writeBuffer2File(buffer *bytes.Buffer) error {
	if f.lastBlockSize > _const.BlockSize {
		panic("lastBlockSize exceeded BlockSize")
	}
	_, err := f.fd.Write(buffer.Bytes())
	return err
}

func (f *SegmentFile) appendChunk2Buffer(buffer *bytes.Buffer, data []byte, cType ChunkType) error {
	// 设置header中的长度
	binary.LittleEndian.PutUint16(f.header[4:6], uint16(len(data)))
	// 设置header中的类型
	f.header[6] = cType

	// 对 len + type + data 求 checksum
	sum := crc32.ChecksumIEEE(f.header[4:])
	sum = crc32.Update(sum, crc32.IEEETable, data)
	// 设置header中的校验和
	binary.LittleEndian.PutUint32(f.header[:4], sum)
	//将一个完整的chunk写入buf中（header + payload 就是一个chunk）
	_, err := buffer.Write(f.header[:])
	if err != nil {
		return err
	}
	_, err = buffer.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func segmentFileName(dir, ext string, id uint32) string {
	return filepath.Join(dir, fmt.Sprintf("%010d"+ext, id))
}

func openSegmentFile(dir string, ext string, id uint32, localCache *lru.Cache[uint32, []byte]) (*SegmentFile, error) {
	path := segmentFileName(dir, ext, id)
	fd, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = fd.Close()
		}
	}()

	stat, err := fd.Stat()
	if err != nil {
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
