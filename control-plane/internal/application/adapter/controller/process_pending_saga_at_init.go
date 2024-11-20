package controller

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/usecase"
)

type ProcessPendingSagaAtInitController struct {
	LookingForPendingSagaAtInitUseCase usecase.LookingForPendingSagaAtInitUseCase
}

func (p *ProcessPendingSagaAtInitController) Run() {
	p.LookingForPendingSagaAtInitUseCase = usecase.LookingForPendingSagaAtInitUseCase{
		FileChunkRepository:       fileChunkRepositoryImpl,
		FileMetadataRepository:    fileMetadataRepositoryImpl,
		FileReplicationRepository: fileReplicationRequestImpl,
		SagaInformationRepository: sagaInformationRepositoryImpl,
	}
	p.LookingForPendingSagaAtInitUseCase.Execute()
}
