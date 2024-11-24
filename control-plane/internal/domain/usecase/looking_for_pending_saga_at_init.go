package usecase

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/repository"
	saga_entity "github.com/Zando74/GopherFS/control-plane/internal/domain/saga/entity"
	saga_repository "github.com/Zando74/GopherFS/control-plane/internal/domain/saga/repository"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/saga/upload_saga"
)

func isEqual(a, b []saga_entity.SagaInformation) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].SagaName != b[i].SagaName || a[i].FileID != b[i].FileID || a[i].StepName != b[i].StepName {
			return false
		}
	}
	return true
}

type LookingForPendingSagaAtInitUseCase struct {
	FileChunkRepository       repository.FileChunkRepository
	FileMetadataRepository    repository.FileMetadataRepository
	FileReplicationRepository repository.FileReplicationRequestRepository
	SagaInformationRepository saga_repository.SagaInformationRepository
	prevSagaInfos             []saga_entity.SagaInformation
}

func (s *LookingForPendingSagaAtInitUseCase) Execute() {
	go func() {

		// Due to the chaotic nature of the election process without quorum, We could switch between leader and follower multiple times.
		// Causing multiple call to RecoverPendingSagas,
		// So we need to check if the pending sagas are already processed or not, by storing the previous one.

		pendingSagas, _ := s.SagaInformationRepository.GetPendingSagaInformation()
		if !isEqual(pendingSagas, s.prevSagaInfos) && len(pendingSagas) != 0 {
			s.prevSagaInfos = pendingSagas

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
		}
	}()
}
