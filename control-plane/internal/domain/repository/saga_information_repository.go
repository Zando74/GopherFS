package repository

import "github.com/Zando74/GopherFS/control-plane/internal/domain/entity"

type SagaInformationRepository interface {
	GetPendingSagaInformation() ([]entity.SagaInformation, error)
	GetFailedSagaInformation() ([]entity.SagaInformation, error)
	MarkSagaStepAsDone(step entity.SagaInformation) error
}
