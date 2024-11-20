package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Zando74/GopherFS/control-plane/internal/application/adapter/consensus"
	"github.com/Zando74/GopherFS/control-plane/internal/application/adapter/consensus/fsm"
	saga_entity "github.com/Zando74/GopherFS/control-plane/internal/domain/saga/entity"
	saga_repository "github.com/Zando74/GopherFS/control-plane/internal/domain/saga/repository"
	"github.com/Zando74/GopherFS/control-plane/logger"
	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
)

type SagaInformationRepositoryImpl struct {
	client *badger.DB
	raft   *raft.Raft
	key    string
}

func NewSagaInformationRepository() saga_repository.SagaInformationRepository {
	badgerDB := fsm.BabdgerSingleton.GetInstance()
	raftsrv := consensus.RaftServerSingleton.GetInstance()

	return &SagaInformationRepositoryImpl{
		client: badgerDB,
		raft:   raftsrv,
		key:    "saga_information",
	}
}

func (repo *SagaInformationRepositoryImpl) getSagaInfos() []saga_entity.SagaInformation {

	txn := repo.client.NewTransaction(false)
	defer txn.Discard()

	item, _ := txn.Get([]byte(repo.key))
	if item == nil {
		return []saga_entity.SagaInformation{}
	}

	value := make([]byte, 0)

	err := item.Value(func(val []byte) error {
		value = val
		return nil
	})

	if err != nil {
		logger.LoggerSingleton.GetInstance().Error(err)
	}

	var sagaInfos []saga_entity.SagaInformation
	if len(value) > 0 {
		json.Unmarshal(value, &sagaInfos)
	}

	return sagaInfos
}

func (repo *SagaInformationRepositoryImpl) SaveSagaInformation(sagaInformation saga_entity.SagaInformation) error {

	sagaInfos := repo.getSagaInfos()
	sagaInfos = append(sagaInfos, sagaInformation)

	payload := fsm.CommandPayload{
		Operation: "SET",
		Key:       repo.key,
		Value:     sagaInfos,
	}

	data, _ := json.Marshal(payload)

	applyFuture := repo.raft.Apply(data, 500*time.Millisecond)

	if err := applyFuture.Error(); err != nil {
		return err
	}
	response := applyFuture.Response()
	if response == nil {
		return fmt.Errorf("raft apply response is nil")
	}
	_, ok := response.(*fsm.ApplyResponse)
	if !ok {
		return fmt.Errorf("invalid raft apply response")
	}

	return nil
}

func (repo *SagaInformationRepositoryImpl) GetPendingSagaInformation() ([]saga_entity.SagaInformation, error) {

	sagaInfos := repo.getSagaInfos()

	var pendingSagaInfos []saga_entity.SagaInformation
	for _, sagaInfo := range sagaInfos {
		if sagaInfo.Status == saga_entity.Pending {
			pendingSagaInfos = append(pendingSagaInfos, sagaInfo)
		}
	}

	return pendingSagaInfos, nil
}

func (repo *SagaInformationRepositoryImpl) GetFailedSagaInformation() ([]saga_entity.SagaInformation, error) {

	sagaInfos := repo.getSagaInfos()

	var failedSagaInfos []saga_entity.SagaInformation
	for _, sagaInfo := range sagaInfos {
		if sagaInfo.Status == saga_entity.Failed {
			failedSagaInfos = append(failedSagaInfos, sagaInfo)
		}
	}

	return failedSagaInfos, nil
}

func (repo *SagaInformationRepositoryImpl) MarkSagaStepAsDone(fileID string) error {
	sagaInfos := repo.getSagaInfos()

	for i, sagaInfo := range sagaInfos {
		if sagaInfo.FileID == fileID {
			sagaInfos = append(sagaInfos[:i], sagaInfos[i+1:]...)
			break
		}
	}
	payload := fsm.CommandPayload{
		Operation: "SET",
		Key:       repo.key,
		Value:     sagaInfos,
	}

	data, _ := json.Marshal(payload)

	applyFuture := repo.raft.Apply(data, 500*time.Millisecond)

	if err := applyFuture.Error(); err != nil {
		return err
	}

	response := applyFuture.Response()
	if response == nil {
		return fmt.Errorf("raft apply response is nil")
	}
	_, ok := response.(*fsm.ApplyResponse)
	if !ok {
		return fmt.Errorf("invalid raft apply response")
	}

	return nil
}

func (repo *SagaInformationRepositoryImpl) MarkSagaStepAsFailed(fileID string) error {

	sagaInfos := repo.getSagaInfos()
	for i, sagaInfo := range sagaInfos {
		if sagaInfo.FileID == fileID {
			sagaInfos = append(sagaInfos[:i], sagaInfos[i+1:]...)
			break
		}
	}

	payload := fsm.CommandPayload{
		Operation: "SET",
		Key:       repo.key,
		Value:     sagaInfos,
	}

	data, _ := json.Marshal(payload)

	applyFuture := repo.raft.Apply(data, 500*time.Millisecond)

	if err := applyFuture.Error(); err != nil {
		return err
	}

	response := applyFuture.Response()
	if response == nil {
		return fmt.Errorf("raft apply response is nil")
	}
	_, ok := response.(*fsm.ApplyResponse)
	if !ok {
		return fmt.Errorf("invalid raft apply response")
	}

	return nil
}
