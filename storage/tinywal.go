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

func (w *TinyWAL) WriteAll() ([]*ChunkPosition, error) {
	if len(w.pendingWrites) == 0 {
		return make([]*ChunkPosition, 0), nil
	}
	w.mutex.Lock()
	defer func() {
		w.ClearPendingWrites()
		w.mutex.Unlock()
	}()

	if w.pendingWritesSize > w.option.MemTableSize {
		return nil, _const.ErrorPendingSizeTooLarge
	}

	if uint64(w.activeSegment.Size())+w.pendingWritesSize > w.option.MemTableSize {
		err := w.replaceActiveSegmentFile()
		if err != nil {
			return nil, err
		}
	}
	all, err := w.activeSegment.WriteAll(w.pendingWrites)
	if err != nil {
		return nil, err
	}
	return all, nil
}

func (w *TinyWAL) Sync() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.activeSegment.Sync()
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
	var readers []*SegmentReader
	for _, segment := range w.immutableSegment {
		readers = append(readers, segment.NewSegmentReader())
	}
	readers = append(readers, w.activeSegment.NewSegmentReader())

	sort.Slice(readers, func(i, j int) bool { return readers[i].seg.segmentFileId < readers[j].seg.segmentFileId })

	return &_const.Reader{
		AllSegmentReader: readers,
		Progress:         0,
	}

}

func (w *TinyWAL) Write(data []byte) (*ChunkPosition, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.maxWriteSize(int64(len(data))) > int64(w.option.MemTableSize) {
		return nil, _const.ErrorDataToLarge
	}

	if w.isFull(int64(len(data))) {
		if err := w.replaceActiveSegmentFile(); err != nil {
			return nil, err
		}
	}
	position, err := w.activeSegment.Write(data)
	if err != nil {
		return nil, err
	}
	w.byteWrite += uint64(position.ChunkSize)

	isSync := w.option.Sync
	if !isSync && w.byteWrite > w.option.BytesPerSync {
		isSync = true
	}
	if isSync {
		err := w.activeSegment.Sync()
		if err != nil {
			return nil, err
		}
		w.byteWrite = 0
	}
	return position, err
}

func (w *TinyWAL) isFull(delta int64) bool {
	return w.activeSegment.Size()+w.maxWriteSize(delta) > int64(w.option.MemTableSize)
}

func (w *TinyWAL) replaceActiveSegmentFile() error {
	err := w.activeSegment.Sync()
	if err != nil {
		return err
	}
	w.byteWrite = 0
	file, err := openSegmentFile(w.option.DirPath, w.option.segmentFileExt, w.activeSegment.segmentFileId+1, w.localCache)
	if err != nil {
		return err
	}
	w.immutableSegment[w.activeSegment.segmentFileId] = w.activeSegment
	w.activeSegment = file
	return nil
}

func (w *TinyWAL) ClearPendingWrites() {

}
