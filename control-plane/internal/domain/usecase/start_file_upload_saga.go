package usecase

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/coordinator"
)

type StartFileUploadSagaUseCase struct {
	UploadSagaCoordinator *coordinator.UploadSagaCoordinator
}

func (s *StartFileUploadSagaUseCase) Execute() {
	go func() {
		s.UploadSagaCoordinator.StartSaga()
	}()
}
