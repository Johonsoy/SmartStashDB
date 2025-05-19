package storage

import (
	_const "SmartStashDB/const"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru/v2"
	"io/fs"
	"os"
	"sort"
	"strings"
	"sync"
)

type TinyWAL struct {
	option           WalOptions
	mutex            sync.RWMutex
	activeSegment    *SegmentFile
	immutableSegment map[SegmentFileId]*SegmentFile
	localCache       *lru.Cache[uint32, []byte]
	byteWrite        uint64

	pendingWritesLock sync.Mutex
	pendingWrites     [][]byte
	pendingWritesSize uint64
}

func (w *TinyWAL) close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.localCache != nil {
		w.localCache.Purge()
	}

	for _, segment := range w.immutableSegment {
		if segment != nil {
			if err := segment.Close(); err != nil {
				return err
			}
		}
	}
	w.immutableSegment = nil
	return w.activeSegment.Close()
}

func OpenTinyWAL(option WalOptions) (*TinyWAL, error) {
	if strings.HasPrefix(option.segmentFileExt, ".") {
		return nil, errors.New(option.segmentFileExt + " is not allowed")
	}

	err := os.MkdirAll(option.DirPath, fs.ModePerm)
	if err != nil {
		return nil, err
	}

	tinyWAL := &TinyWAL{
		option:           option,
		immutableSegment: make(map[SegmentFileId]*SegmentFile),
		activeSegment:    nil,
	}

	if option.BlockCache > 0 {
		blockNum := option.BlockCache / _const.BlockSize
		if option.BlockCache%_const.BlockSize != 0 {
			blockNum++
		}
		tinyWAL.localCache, err = lru.New[uint32, []byte](int(blockNum))
		if err != nil {
			return nil, err
		}
	}

	dir, err := os.ReadDir(option.DirPath)

	if err != nil {
		return nil, err
	}

	var segmentFileIds []int
	for _, file := range dir {
		if file.IsDir() {
			continue
		}
		segmentFileId := 0
		_, err := fmt.Scanf(file.Name(), "%d"+option.segmentFileExt, &segmentFileId)
		if err != nil {
			continue
		}
		segmentFileIds = append(segmentFileIds, segmentFileId)
	}

	if len(segmentFileIds) == 0 {
		segment, err := openSegmentFile(option.DirPath, option.segmentFileExt, _const.FirstSegmentFileId, tinyWAL.localCache)

		if err != nil {
			return nil, err
		}
		tinyWAL.activeSegment = segment
	} else {
		sort.Ints(segmentFileIds)
		for i, fileId := range segmentFileIds {
			segment, err := openSegmentFile(option.DirPath, option.segmentFileExt, uint32(fileId), tinyWAL.localCache)
			if err != nil {
				return nil, err
			}
			if i == len(segmentFileIds)-1 {
				tinyWAL.activeSegment = segment
			} else {
				tinyWAL.immutableSegment[uint32(fileId)] = segment
			}
		}
	}

	return tinyWAL, nil
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

func (w *TinyWAL) NewReader() *_const.Reader {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	var readers []*_const.Reader
	for _, segment := range w.immutableSegment {
		readers = append(readers, segment.NewSegmentReader())
	}
}
