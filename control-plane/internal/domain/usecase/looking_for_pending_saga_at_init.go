package usecase

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/repository"
	saga_entity "github.com/Zando74/GopherFS/control-plane/internal/domain/saga/entity"
	saga_repository "github.com/Zando74/GopherFS/control-plane/internal/domain/saga/repository"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/saga/upload_saga"
)

type LookingForPendingSagaAtInitUseCase struct {
	FileChunkRepository       repository.FileChunkRepository
	FileMetadataRepository    repository.FileMetadataRepository
	FileReplicationRepository repository.FileReplicationRequestRepository
	SagaInformationRepository saga_repository.SagaInformationRepository
}

func (s *LookingForPendingSagaAtInitUseCase) Execute() {
	go func() {

		pendingSagas, err := s.SagaInformationRepository.GetPendingSagaInformation()
		if err != nil {
			return
		}

		for _, saga := range pendingSagas {
			switch saga.SagaName {
			case saga_entity.UploadSaga:
				fileMetadata := &entity.FileMetadata{
					FileID:   saga.FileID,
					FileName: "",
					FileSize: 0,
					Chunks:   []entity.ChunkLocations{},
				}

				switch saga.StepName {
				case saga_entity.FileChunkSave:
					fileChunkSaveStep := upload_saga.NewFileChunkSaveConfirmation(
						s.FileChunkRepository, fileMetadata, s.SagaInformationRepository,
					)
					fileChunkSaveStep.Rollback()
				case saga_entity.FileMetadataSave:

					/**
					Instead of rollback here, we should define a way to rebuild metadata
					by calling a function that will connect to each storage service to get all chunks locations
					asynchronously and then save the metadata. and Process next step. Same as the way we store chunks
					#NOTIMPLEMENTED
					*/

					fileChunkSaveStep := upload_saga.NewFileChunkSaveConfirmation(
						s.FileChunkRepository, fileMetadata, s.SagaInformationRepository,
					)
					fileSaveMetadataStep := upload_saga.NewFileSaveMetadata(s.FileMetadataRepository, fileMetadata, s.SagaInformationRepository)
					fileSaveMetadataStep.Rollback()
					fileChunkSaveStep.Rollback()
				case saga_entity.ReplicateFile:
					fileReplicationStep := upload_saga.NewFileReplication(s.FileReplicationRepository, &entity.FileReplicationRequest{FileID: saga.FileID}, s.SagaInformationRepository)
					fileReplicationStep.Transaction()
				}
			}

		}

	}()
}
