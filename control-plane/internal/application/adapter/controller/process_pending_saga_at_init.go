package controller

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/usecase"
)

type ProcessPendingSagaAtInitController struct {
	LookingForPendingSagaAtInitUseCase usecase.LookingForPendingSagaAtInitUseCase
}

func NewProcessPendingSagaAtInitController() *ProcessPendingSagaAtInitController {
	return &ProcessPendingSagaAtInitController{
		LookingForPendingSagaAtInitUseCase: usecase.LookingForPendingSagaAtInitUseCase{
			FileChunkRepository:       fileChunkRepositoryImpl,
			FileMetadataRepository:    fileMetadataRepositoryImpl,
			FileReplicationRepository: fileReplicationRequestImpl,
			SagaInformationRepository: sagaInformationRepositoryImpl,
		},
	}
}

func (p *ProcessPendingSagaAtInitController) Run() {
	p.LookingForPendingSagaAtInitUseCase.Execute()
}
