package coordinator

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/repository"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/saga/upload_saga"
)

type UploadSagaCoordinator struct {
	Saga                          *entity.Saga
	fileChunkSaveConfirmationStep *upload_saga.FileChunkSaveConfirmation
	fileSaveMetadataStep          *upload_saga.FileSaveMetadata
}

func NewUploadSagaCoordinator(
	fileChunkRepository repository.FileChunkRepository,
	fileMetadataRepository repository.FileMetadataRepository,
	fileID, fileName string,
) *UploadSagaCoordinator {
	fileMetadata := &entity.FileMetadata{
		FileID:   fileID,
		FileName: fileName,
		FileSize: 0,
		Chunks:   []entity.ChunkLocations{},
	}

	fileChunkSaveConfirmationStep := upload_saga.NewFileChunkSaveConfirmation(fileChunkRepository, fileMetadata)
	fileSaveMetadataStep := upload_saga.NewFileSaveMetadata(fileMetadataRepository, fileMetadata)

	uploadSaga := &entity.Saga{
		Steps: []entity.SagaStep{
			fileChunkSaveConfirmationStep,
			fileSaveMetadataStep,
		},
	}
	return &UploadSagaCoordinator{
		Saga:                          uploadSaga,
		fileChunkSaveConfirmationStep: fileChunkSaveConfirmationStep,
		fileSaveMetadataStep:          fileSaveMetadataStep,
	}
}

func (usc *UploadSagaCoordinator) StartSaga() {
	usc.Saga.Execute()
}

func (usc *UploadSagaCoordinator) RollbackSaga() {
	for i := range usc.Saga.Steps {
		step := usc.Saga.Steps[i]
		if step.IsStarted() {
			step.Fail()
			break
		}
	}
}

func (usc *UploadSagaCoordinator) CommitFileChunk(fileChunk *entity.FileChunk) error {
	return usc.fileChunkSaveConfirmationStep.ProcessChunk(*fileChunk)
}

func (usc *UploadSagaCoordinator) StartWaitingForConfirmations() {
	usc.fileChunkSaveConfirmationStep.StartWaitingForConfirmations()
}
