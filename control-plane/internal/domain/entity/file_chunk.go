package entity

import "github.com/Zando74/GopherFS/control-plane/internal/domain/valueobject"

type FileChunk struct {
	Filename       string
	SequenceNumber string
	ChunkID        valueobject.ChunkHash
	ChunkData      valueobject.Chunk
}
