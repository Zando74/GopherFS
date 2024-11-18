package mock_repository

import (
	"time"

	"math/rand"

	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
)

type FileChunkRepositoryMock struct {
	files []*entity.FileChunk
}

func (fcrm *FileChunkRepositoryMock) SaveFileChunk(chunk *entity.FileChunk, onSuccess func(fileChunkLocation entity.ChunkLocations)) error {
	fcrm.files = append(fcrm.files, chunk)
	go func() {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		onSuccess(entity.ChunkLocations{
			ChunkID:         chunk.FileID,
			SequenceNumber:  chunk.SequenceNumber,
			PrimaryLocation: entity.Location{StorageID: "1"},
		})
	}()

	return nil
}

func (fcrm *FileChunkRepositoryMock) DeleteAllFilesChunks(fileID string) error {
	fcrm.files = nil
	return nil
}

func (fcrm *FileChunkRepositoryMock) GetFiles() []*entity.FileChunk {
	return fcrm.files
}
