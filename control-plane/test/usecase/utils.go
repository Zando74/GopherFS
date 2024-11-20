package test

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/saga/coordinator"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/usecase"
	mock_repository "github.com/Zando74/GopherFS/control-plane/test/mock/repository"
)

func GenerateFileUploadSagaUseCase(fileID, fileName string) (
	*usecase.SplitBatchToChunkUseCase,
	*coordinator.UploadSagaCoordinator,
	*mock_repository.FileChunkRepositoryMock,
) {
	var fileChunkTestRepository = &mock_repository.FileChunkRepositoryMock{}
	var fileMetadataRepository = &mock_repository.FileMetadataRepository{}
	var fileReplicationRequestRepository = &mock_repository.FileReplicationRequestRepository{}
	var sagaInformationRepository = &mock_repository.SagaInformationRepositoryImpl{}

	var uploadSagaCoordinator = coordinator.NewUploadSagaCoordinator(
		fileChunkTestRepository,
		fileMetadataRepository,
		fileReplicationRequestRepository,
		sagaInformationRepository,
		fileID,
		fileName,
	)

	var splitBatchToChunkUseCase = &usecase.SplitBatchToChunkUseCase{
		UploadSagaCoordinator: uploadSagaCoordinator,
	}

	return splitBatchToChunkUseCase, uploadSagaCoordinator, fileChunkTestRepository
}
