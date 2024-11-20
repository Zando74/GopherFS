package saga_repository

import saga_entity "github.com/Zando74/GopherFS/control-plane/internal/domain/saga/entity"

type SagaInformationRepository interface {
	SaveSagaInformation(sagaInformation saga_entity.SagaInformation) error
	GetPendingSagaInformation() ([]saga_entity.SagaInformation, error)
	GetFailedSagaInformation() ([]saga_entity.SagaInformation, error)
	MarkSagaStepAsDone(fileID string) error
	MarkSagaStepAsFailed(fileID string) error
}
