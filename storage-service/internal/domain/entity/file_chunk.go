package entity

import "github.com/Zando74/GopherFS/storage-service/internal/domain/value_object"

type FileChunk struct {
	FileID         string
	SequenceNumber uint32
	ChunkID        value_object.ChunkHash
	ChunkData      value_object.Chunk
}
