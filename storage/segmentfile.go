package storage

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru/v2"
	"os"
	"path/filepath"
)

type SegmentFileId = uint32

type segmentFile struct {
	segmentFileId SegmentFileId

	fd *os.File

	lastBlockIndex uint32

	lastBlockSize uint32

	header []byte

	closed bool

	localCache *lru.Cache[uint32, []byte]
}

func segmentFileName(dir, ext string, id uint32) string {
	return filepath.Join(dir, fmt.Sprintf("%010d"+ext, id))
}

func openSegmentFile(dir string, ext string, id uint32, localCache *lru.Cache[uint32, []byte]) {
	os.OpenFile(segmentFileName(dir, ext, id), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
}
