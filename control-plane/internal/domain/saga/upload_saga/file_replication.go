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

type FileReplication struct {
	saga.ConcreteSagaStep
	FileReplicationRequestSaver repository.FileReplicationRequestRepository
	FileReplicationRequest      *entity.FileReplicationRequest
}

func NewFileReplication(fileReplicationRequestSaver repository.FileReplicationRequestRepository, fileReplicationRequest *entity.FileReplicationRequest, sagaInformationRepository saga_repository.SagaInformationRepository) *FileReplication {
	return &FileReplication{
		FileReplicationRequestSaver: fileReplicationRequestSaver,
		FileReplicationRequest:      fileReplicationRequest,
		ConcreteSagaStep: saga.ConcreteSagaStep{
			SagaInformationLogger: sagaInformationRepository,
			StatusChannel:         make(chan saga.SagaState),
		},
	}
}

func (f *FileReplication) Execute() error {
	f.ExpectConfirmation()
	f.StartWaitingForConfirmations()
	err := f.FileReplicationRequestSaver.SaveReplicationRequest(f.FileReplicationRequest, f.Ack)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileReplication) Transaction() error {
	f.SetTTL(config.ConfigSingleton.GetInstance().FileStorage.Saga_ttl)
	log.Info(logger.StartingFileUploadSagaStepMessage, saga_entity.ReplicateFile, f.FileReplicationRequest.FileID)

	f.SagaInformationLogger.SaveSagaInformation(saga_entity.SagaInformation{
		FileID:           f.FileReplicationRequest.FileID,
		StepName:         saga_entity.ReplicateFile,
		SagaName:         saga_entity.UploadSaga,
		CreatedAt:        time.Now().Unix(),
		Status:           saga_entity.Pending,
		RecoveryStrategy: saga_entity.RollBack,
	})

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
				log.Info(logger.FailedFileUploadSagaStepMessage, saga_entity.ReplicateFile, f.FileReplicationRequest.FileID)
				f.SagaInformationLogger.MarkSagaStepAsFailed(f.FileReplicationRequest.FileID)
				return fmt.Errorf("Canceled")
			}

		default:
			if f.IsWaitingForConfirmations() {
				if f.NoPendingConfirmation() {
					log.Info(logger.FinishedFileUploadSagaStepMessage, saga_entity.ReplicateFile, f.FileReplicationRequest.FileID)
					f.SagaInformationLogger.MarkSagaStepAsDone(f.FileReplicationRequest.FileID)
					return nil
				}
			}
		}

	}

}

func (f *FileReplication) Rollback() error {
	log.Info(logger.RecoveringFileUploadSagaStepSagaMessage, saga_entity.ReplicateFile, f.FileReplicationRequest.FileID)
	f.FileReplicationRequestSaver.DeleteReplicationRequest(f.FileReplicationRequest)
	f.SagaInformationLogger.MarkSagaStepAsFailed(f.FileReplicationRequest.FileID)
	return nil
}
