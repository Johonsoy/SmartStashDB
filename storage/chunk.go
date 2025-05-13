package storage

type ChunkPosition struct {
	SegmentFileId SegmentFileId
	BlockIndex    uint32
	ChunkOffset   uint32
	ChunkSize     uint32
}
