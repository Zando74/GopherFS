package mock_repository

import (
	"time"

	"math/rand"

	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
)

type FileChunkRepositoryMock struct {
	Files   []*entity.FileChunk
	Deleted bool
}

func (fcrm *FileChunkRepositoryMock) SaveFileChunk(chunk *entity.FileChunk, onSuccess func(fileChunkLocation entity.ChunkLocations)) error {
	fcrm.Files = append(fcrm.Files, chunk)
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
	fcrm.Files = nil
	fcrm.Deleted = true
	return nil
}

func (fcrm *FileChunkRepositoryMock) GetFiles() []*entity.FileChunk {
	if fcrm.Deleted {
		return nil
	}
	return fcrm.Files
}
