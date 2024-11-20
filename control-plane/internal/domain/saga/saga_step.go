package saga

import (
	"sync"

	saga_repository "github.com/Zando74/GopherFS/control-plane/internal/domain/saga/repository"
)

type Confirmation byte

type SagaStep interface {
	Transaction() error
	Rollback() error
	ExpectConfirmation()
	StartWaitingForConfirmations()
	IsStarted() bool
	IsWaitingForConfirmations() bool
	IsFailed() bool
	NoPendingConfirmation() bool
	SetTTL(ttl uint32)
	DecreaseTTL()
	IsOverTTL() bool
	Fail()
}

type ConcreteSagaStep struct {
	Status                SagaState
	ConfirmationsPool     []Confirmation
	Mu                    sync.Mutex
	StatusChannel         chan SagaState
	TTL                   uint32
	SagaInformationLogger saga_repository.SagaInformationRepository
}

func (s *ConcreteSagaStep) ExpectConfirmation() {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.ConfirmationsPool = append(s.ConfirmationsPool, Confirmation(0))
}

func (s *ConcreteSagaStep) Ack() {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.ConfirmationsPool = s.ConfirmationsPool[1:]
}

func (s *ConcreteSagaStep) StartWaitingForConfirmations() {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Status = WaitingForConfirmation
	go func() {
		s.StatusChannel <- WaitingForConfirmation
	}()

}

func (s *ConcreteSagaStep) IsStarted() bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return s.Status >= Pending && s.Status < Done
}

func (s *ConcreteSagaStep) IsWaitingForConfirmations() bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return s.Status == WaitingForConfirmation
}

func (s *ConcreteSagaStep) IsFailed() bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return s.Status == Failed
}

func (s *ConcreteSagaStep) NoPendingConfirmation() bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return len(s.ConfirmationsPool) == 0
}

func (s *ConcreteSagaStep) SetTTL(ttl uint32) {
	s.TTL = ttl
}

func (s *ConcreteSagaStep) DecreaseTTL() {
	s.TTL--
	if s.TTL <= 0 {
		s.Fail()
	}
}

func (s *ConcreteSagaStep) IsOverTTL() bool {
	return s.TTL == 0
}

func (s *ConcreteSagaStep) Fail() {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Status = Failed

	go func() {
		s.StatusChannel <- Failed
	}()

}
