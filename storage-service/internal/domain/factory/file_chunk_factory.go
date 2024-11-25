package factory

import (
	"github.com/Zando74/GopherFS/storage-service/internal/domain/entity"
	"github.com/Zando74/GopherFS/storage-service/internal/domain/value_object"
)

type FileChunkInput struct {
	FileID         string
	SequenceNumber uint32
	StableContext  string
	ChunkData      []byte
}

type FileChunkFactory struct{}

func (fc FileChunkFactory) CreateFileChunk(input FileChunkInput) (*entity.FileChunk, error) {
	chunk, err := value_object.NewChunk(input.ChunkData)
	if err != nil {
		return nil, err
	}
	chunkHash := value_object.NewChunkHash(*chunk, input.StableContext)
	return &entity.FileChunk{
		FileID:         input.FileID,
		SequenceNumber: input.SequenceNumber,
		ChunkID:        chunkHash,
		ChunkData:      *chunk,
	}, nil
}