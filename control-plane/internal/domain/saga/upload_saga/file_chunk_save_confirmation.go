package upload_saga

import (
	"fmt"
	"time"

	"github.com/Zando74/GopherFS/control-plane/config"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/repository"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/saga"
	saga_entity "github.com/Zando74/GopherFS/control-plane/internal/domain/saga/entity"
	saga_repository "github.com/Zando74/GopherFS/control-plane/internal/domain/saga/repository"
	"github.com/Zando74/GopherFS/control-plane/logger"
)

var (
	log = logger.LoggerSingleton.GetInstance()
)

type FileChunkSaveConfirmation struct {
	saga.ConcreteSagaStep
	FileChunkSaver repository.FileChunkRepository
	FileMetadata   *entity.FileMetadata
}

func NewFileChunkSaveConfirmation(fileChunkSaver repository.FileChunkRepository, fileMetadata *entity.FileMetadata, sagaInformationRepository saga_repository.SagaInformationRepository) *FileChunkSaveConfirmation {
	return &FileChunkSaveConfirmation{
		FileChunkSaver: fileChunkSaver,
		FileMetadata:   fileMetadata,
		ConcreteSagaStep: saga.ConcreteSagaStep{
			SagaInformationLogger: sagaInformationRepository,
			StatusChannel:         make(chan saga.SagaState),
		},
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
	if f.FileMetadata.FileSize == 0 {
		f.SagaInformationLogger.SaveSagaInformation(saga_entity.SagaInformation{
			FileID:           f.FileMetadata.FileID,
			StepName:         saga_entity.FileChunkSave,
			SagaName:         saga_entity.UploadSaga,
			CreatedAt:        time.Now().Unix(),
			Status:           saga_entity.Pending,
			RecoveryStrategy: saga_entity.RollBack,
		})
	}
	f.FileMetadata.FileSize += int64(len(fileChunk.ChunkData))
	f.Mu.Unlock()
	f.ConcreteSagaStep.ExpectConfirmation()

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

	log.Info(logger.StartingFileUploadSagaStepMessage, saga_entity.FileChunkSave, f.FileMetadata.FileID)
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			f.DecreaseTTL()
		case status := <-f.StatusChannel:
			if status == saga.Failed {
				log.Info(logger.FailedFileUploadSagaStepMessage, saga_entity.FileChunkSave, f.FileMetadata.FileID)
				f.SagaInformationLogger.MarkSagaStepAsFailed(f.FileMetadata.FileID)
				return fmt.Errorf("Canceled")
			}
		default:
			if f.IsWaitingForConfirmations() {
				if f.NoPendingConfirmation() {
					log.Info(logger.FinishedFileUploadSagaStepMessage, saga_entity.FileChunkSave, f.FileMetadata.FileID)
					f.SagaInformationLogger.MarkSagaStepAsDone(f.FileMetadata.FileID)
					return nil
				}
			}
		}
	}
}

func (f *FileChunkSaveConfirmation) Rollback() error {
	log.Info(logger.RecoveringFileUploadSagaStepSagaMessage, saga_entity.FileChunkSave, f.FileMetadata.FileID)
	f.FileChunkSaver.DeleteAllFilesChunks(f.FileMetadata.FileID)
	f.SagaInformationLogger.MarkSagaStepAsFailed(f.FileMetadata.FileID)
	return nil
}
