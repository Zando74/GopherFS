package repository

import "github.com/Zando74/GopherFS/control-plane/internal/domain/entity"

type SagaCoordinator interface {
	StartSaga()
	RollbackSaga()
	CommitFileChunk(fileChunk *entity.FileChunk) error
	StartWaitingForConfirmations()
}
