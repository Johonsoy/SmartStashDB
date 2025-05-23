package storage

type ChunkType = byte

const (
	ChunkTypeFull ChunkType = iota
	ChunkTypeStart
	ChunkTypeMiddle
	ChunkTypeEnd
)

type ChunkPosition struct {
	SegmentFileId SegmentFileId
	BlockIndex    uint32
	ChunkOffset   uint32
	ChunkSize     uint32
}
