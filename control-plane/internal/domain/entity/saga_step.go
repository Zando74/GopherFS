package entity

import (
	"sync"
	"time"
)

type Confirmation byte

type SagaStep interface {
	Transaction() error
	Rollback() error
	Start()
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
	Status            SagaState
	ConfirmationsPool []Confirmation
	Mu                sync.Mutex
	TTL               uint32
}

func (s *ConcreteSagaStep) Start() {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Status = Pending
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
	time.Sleep(1 * time.Second)
	s.TTL--
}

func (s *ConcreteSagaStep) IsOverTTL() bool {
	return s.TTL == 0
}

func (s *ConcreteSagaStep) Fail() {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Status = Failed
}
