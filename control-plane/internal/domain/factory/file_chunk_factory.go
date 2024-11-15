package factory

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/valueobject"
)

type FileChunkInput struct {
	Filename       string
	SequenceNumber string
	StableContext  string
	ChunkData      []byte
}

type FileChunkFactory struct{}

func (fc FileChunkFactory) CreateFileChunk(input FileChunkInput) (*entity.FileChunk, error) {
	chunk, err := valueobject.NewChunk(input.ChunkData)
	if err != nil {
		return nil, err
	}
	chunkHash := valueobject.NewChunkHash(*chunk, input.StableContext)
	return &entity.FileChunk{
		Filename:       input.Filename,
		SequenceNumber: input.SequenceNumber,
		ChunkID:        chunkHash,
		ChunkData:      *chunk,
	}, nil
}
