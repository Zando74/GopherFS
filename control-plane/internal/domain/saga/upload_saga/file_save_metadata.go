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

type FileSaveMetadata struct {
	saga.ConcreteSagaStep
	FileMetadataSaver repository.FileMetadataRepository
	FileMetadata      *entity.FileMetadata
}

func NewFileSaveMetadata(fileMetadataSaver repository.FileMetadataRepository, fileMetadata *entity.FileMetadata, sagaInformationRepository saga_repository.SagaInformationRepository) *FileSaveMetadata {
	return &FileSaveMetadata{
		FileMetadataSaver: fileMetadataSaver,
		FileMetadata:      fileMetadata,
		ConcreteSagaStep: saga.ConcreteSagaStep{
			SagaInformationLogger: sagaInformationRepository,
			StatusChannel:         make(chan saga.SagaState),
		},
	}
}

func (f *FileSaveMetadata) Execute() error {
	f.SagaInformationLogger.SaveSagaInformation(saga_entity.SagaInformation{
		FileID:           f.FileMetadata.FileID,
		StepName:         saga_entity.FileMetadataSave,
		SagaName:         saga_entity.UploadSaga,
		CreatedAt:        time.Now().Unix(),
		Status:           saga_entity.Pending,
		RecoveryStrategy: saga_entity.RollBack,
	})
	f.ExpectConfirmation()
	f.StartWaitingForConfirmations()
	err := f.FileMetadataSaver.SaveFileMetadata(f.FileMetadata, f.Ack)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSaveMetadata) Transaction() error {
	f.SetTTL(config.ConfigSingleton.GetInstance().FileStorage.Saga_ttl)
	log := logger.LoggerSingleton.GetInstance()
	log.Info(logger.StartingFileUploadSagaStepMessage, saga_entity.FileMetadataSave, f.FileMetadata.FileID)

	err := f.Execute()

	if err != nil {
		f.Fail()
	}

	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			f.DecreaseTTL()

		case status := <-f.StatusChannel:
			if status == saga.Failed {
				log.Info(logger.FailedFileUploadSagaStepMessage, saga_entity.FileMetadataSave, f.FileMetadata.FileID)
				f.SagaInformationLogger.MarkSagaStepAsFailed(f.FileMetadata.FileID)
				return fmt.Errorf("Canceled")
			}

		default:
			if f.IsWaitingForConfirmations() {
				if f.NoPendingConfirmation() {
					log.Info(logger.FinishedFileUploadSagaStepMessage, saga_entity.FileMetadataSave, f.FileMetadata.FileID)
					f.SagaInformationLogger.MarkSagaStepAsDone(f.FileMetadata.FileID)
					return nil
				}
			}
		}
	}

}

func (f *FileSaveMetadata) Rollback() error {
	log.Info(logger.RecoveringFileUploadSagaStepSagaMessage, saga_entity.FileMetadataSave, f.FileMetadata.FileID)
	f.FileMetadataSaver.DeleteFileMetadata(f.FileMetadata.FileID)
	f.SagaInformationLogger.MarkSagaStepAsFailed(f.FileMetadata.FileID)
	return nil
}
