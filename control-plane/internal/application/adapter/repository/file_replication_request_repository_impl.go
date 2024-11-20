package repository

import (
	"math/rand"
	"time"

	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
)

type FileReplicationRequestRepository struct{}

func (fcr FileReplicationRequestRepository) SaveReplicationRequest(replicationRequest *entity.FileReplicationRequest, onSuccess func()) error {

	go func() {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		onSuccess()
	}()

	return nil
}

func (fcr FileReplicationRequestRepository) DeleteReplicationRequest(replicationRequest *entity.FileReplicationRequest) error {
	return nil
}
