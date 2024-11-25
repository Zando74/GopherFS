package entity

import "github.com/Zando74/GopherFS/control-plane/internal/domain/value_object"

type FileChunk struct {
	FileID         string
	SequenceNumber int
	ChunkID        value_object.ChunkHash
	ChunkData      value_object.Chunk
}
