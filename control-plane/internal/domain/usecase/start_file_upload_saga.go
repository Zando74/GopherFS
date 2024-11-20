package usecase

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/repository"
)

type StartFileUploadSagaUseCase struct {
	UploadSagaCoordinator repository.SagaCoordinator
}

func (s *StartFileUploadSagaUseCase) Execute() {

	go func() {
		s.UploadSagaCoordinator.StartSaga()
	}()
}

func (s *StartFileUploadSagaUseCase) ExecuteSync() {
	s.UploadSagaCoordinator.StartSaga()
}
