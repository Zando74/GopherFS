package mock_repository

import (
	"errors"
	"sync"

	saga_entity "github.com/Zando74/GopherFS/control-plane/internal/domain/saga/entity"
	saga_repository "github.com/Zando74/GopherFS/control-plane/internal/domain/saga/repository"
)

type SagaInformationRepositoryImpl struct {
	mu    sync.Mutex
	sagas []saga_entity.SagaInformation
}

func NewSagaInformationRepository() saga_repository.SagaInformationRepository {
	return &SagaInformationRepositoryImpl{
		sagas: []saga_entity.SagaInformation{},
	}
}

func (repo *SagaInformationRepositoryImpl) SaveSagaInformation(sagaInformation saga_entity.SagaInformation) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, saga := range repo.sagas {
		if saga.FileID == sagaInformation.FileID {
			return errors.New("saga information already exists")
		}
	}

	repo.sagas = append(repo.sagas, sagaInformation)
	return nil
}

func (repo *SagaInformationRepositoryImpl) GetPendingSagaInformation() ([]saga_entity.SagaInformation, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	var pendingSagas []saga_entity.SagaInformation
	for _, saga := range repo.sagas {
		if saga.Status == saga_entity.Pending {
			pendingSagas = append(pendingSagas, saga)
		}
	}
	return pendingSagas, nil
}

func (repo *SagaInformationRepositoryImpl) GetFailedSagaInformation() ([]saga_entity.SagaInformation, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	var failedSagas []saga_entity.SagaInformation
	for _, saga := range repo.sagas {
		if saga.Status == saga_entity.Failed {
			failedSagas = append(failedSagas, saga)
		}
	}
	return failedSagas, nil
}

func (repo *SagaInformationRepositoryImpl) MarkSagaStepAsDone(fileID string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for i, saga := range repo.sagas {
		if saga.FileID == fileID {
			repo.sagas = append(repo.sagas[:i], repo.sagas[i+1:]...)
			return nil
		}
	}

	return errors.New("saga step not found")
}

func (repo *SagaInformationRepositoryImpl) MarkSagaStepAsFailed(fileID string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for i, saga := range repo.sagas {
		if saga.FileID == fileID {
			repo.sagas[i].Status = saga_entity.Failed
			return nil
		}
	}

	return errors.New("saga step not found")
}
