package tinywal

type ChunkType = byte

const (
	ChunkTypeFull ChunkType = iota
	ChunkTypeStart
	ChunkTypeMiddle
	ChunkTypeEnd
)

type Chunk struct {
	SegmentFileId SegmentFileId

	BlockIndex uint32

	ChunkOffset uint32

	ChunkSize uint32
}
