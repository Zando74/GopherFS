package upload_saga

import (
	"fmt"

	"github.com/Zando74/GopherFS/control-plane/config"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/repository"
	"github.com/Zando74/GopherFS/control-plane/logger"
)

type FileChunkSaveConfirmation struct {
	entity.ConcreteSagaStep
	FileChunkSaver repository.FileChunkRepository
	FileMetadata   *entity.FileMetadata
}

func NewFileChunkSaveConfirmation(fileChunkSaver repository.FileChunkRepository, fileMetadata *entity.FileMetadata) *FileChunkSaveConfirmation {
	return &FileChunkSaveConfirmation{
		FileChunkSaver: fileChunkSaver,
		FileMetadata:   fileMetadata,
	}
}

func (f *FileChunkSaveConfirmation) ProcessChunk(fileChunk entity.FileChunk) error {

	err := f.FileChunkSaver.SaveFileChunk(&fileChunk,
		func(fileChunkLocation entity.ChunkLocations) {
			f.UpdateChunkLocationInMetadata(fileChunkLocation)
		})

	if err != nil {
		return err
	}

	f.Mu.Lock()
	f.FileMetadata.FileSize += int64(len(fileChunk.ChunkData))
	f.Mu.Unlock()
	f.ConcreteSagaStep.ExpectConfirmation()

	logger := logger.LoggerSingleton.GetInstance()
	logger.Debug("Processed chunk: ", fileChunk.SequenceNumber, " Current Size processed: ", f.FileMetadata.FileSize)

	return nil
}

func (f *FileChunkSaveConfirmation) UpdateChunkLocationInMetadata(fileChunkLocation entity.ChunkLocations) {
	f.Mu.Lock()
	f.FileMetadata.Chunks = append(f.FileMetadata.Chunks, fileChunkLocation)
	f.Mu.Unlock()
	f.Ack()
}

func (f *FileChunkSaveConfirmation) Transaction() error {
	f.SetTTL(config.ConfigSingleton.GetInstance().FileStorage.Saga_ttl)
	f.Start()
	for {
		f.DecreaseTTL()
		if f.IsWaitingForConfirmations() {
			if f.NoPendingConfirmation() {
				return nil
			}
		}

		if f.IsFailed() {
			return fmt.Errorf("Canceled")
		}

		if f.IsOverTTL() {
			return fmt.Errorf("TTL expired")
		}
	}
}

func (f *FileChunkSaveConfirmation) Rollback() error {
	f.FileChunkSaver.DeleteAllFilesChunks(f.FileMetadata.FileID)
	return nil
}
