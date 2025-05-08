package storage

type SegmentReader struct {
	seg         *segmentFile
	blockidx    uint32
	chunkoffset uint32
}
