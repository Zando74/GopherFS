package repository

import (
	"math/rand"
	"time"

	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
)

type FileChunkRepository struct{}

func (fcr FileChunkRepository) SaveFileChunk(chunk *entity.FileChunk, onSuccess func(fileChunkLocation entity.ChunkLocations)) error {

	go func() {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		onSuccess(entity.ChunkLocations{
			ChunkID:         string(chunk.ChunkID),
			SequenceNumber:  chunk.SequenceNumber,
			PrimaryLocation: entity.Location{StorageID: "1"},
		})
	}()

	return nil
}

func (fcr FileChunkRepository) DeleteAllFilesChunks(fileID string) error {
	return nil
}
