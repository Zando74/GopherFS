package coordinator

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/repository"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/saga"
	saga_repository "github.com/Zando74/GopherFS/control-plane/internal/domain/saga/repository"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/saga/upload_saga"
)

type UploadSagaCoordinator struct {
	Saga                          *saga.Saga
	FileChunkSaveConfirmationStep *upload_saga.FileChunkSaveConfirmation
	FileSaveMetadataStep          *upload_saga.FileSaveMetadata
	FileReplicationStep           *upload_saga.FileReplication
}

func NewUploadSagaCoordinator(
	fileChunkRepository repository.FileChunkRepository,
	fileMetadataRepository repository.FileMetadataRepository,
	fileReplicationRepository repository.FileReplicationRequestRepository,
	sagaInformationRepository saga_repository.SagaInformationRepository,
	fileID, fileName string,
) *UploadSagaCoordinator {
	fileMetadata := &entity.FileMetadata{
		FileID:   fileID,
		FileName: fileName,
		FileSize: 0,
		Chunks:   []entity.ChunkLocations{},
	}

	fileChunkSaveConfirmationStep := upload_saga.NewFileChunkSaveConfirmation(fileChunkRepository, fileMetadata, sagaInformationRepository)
	fileSaveMetadataStep := upload_saga.NewFileSaveMetadata(fileMetadataRepository, fileMetadata, sagaInformationRepository)
	fileReplicationStep := upload_saga.NewFileReplication(fileReplicationRepository, &entity.FileReplicationRequest{FileID: fileID}, sagaInformationRepository)

	uploadSaga := &saga.Saga{
		Steps: []saga.SagaStep{
			fileChunkSaveConfirmationStep,
			fileSaveMetadataStep,
			fileReplicationStep,
		},
	}
	return &UploadSagaCoordinator{
		Saga:                          uploadSaga,
		FileChunkSaveConfirmationStep: fileChunkSaveConfirmationStep,
		FileSaveMetadataStep:          fileSaveMetadataStep,
		FileReplicationStep:           fileReplicationStep,
	}
}

func (usc *UploadSagaCoordinator) StartSaga() {
	usc.Saga.Execute()
}

func (usc *UploadSagaCoordinator) RollbackSaga() {
	for i := range usc.Saga.Steps {
		step := usc.Saga.Steps[i]
		step.Fail()
	}
}

func (usc *UploadSagaCoordinator) CommitFileChunk(fileChunk *entity.FileChunk) error {
	return usc.FileChunkSaveConfirmationStep.ProcessChunk(*fileChunk)
}

func (usc *UploadSagaCoordinator) StartWaitingForConfirmations() {
	usc.FileChunkSaveConfirmationStep.StartWaitingForConfirmations()
}
