package mock_repository

import (
	"math/rand"
	"time"

	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
)

type FileReplicationRequestRepository struct {
	replicationRequestStorage []*entity.FileReplicationRequest
}

func (fcr *FileReplicationRequestRepository) SaveReplicationRequest(replicationRequest *entity.FileReplicationRequest, onSuccess func()) error {
	fcr.replicationRequestStorage = append(fcr.replicationRequestStorage, replicationRequest)
	go func() {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		onSuccess()
	}()

	return nil
}

func (fcr *FileReplicationRequestRepository) DeleteReplicationRequest(replicationRequest *entity.FileReplicationRequest) error {
	for i, request := range fcr.replicationRequestStorage {
		if request.FileID == replicationRequest.FileID {
			fcr.replicationRequestStorage = append(fcr.replicationRequestStorage[:i], fcr.replicationRequestStorage[i+1:]...)
			break
		}
	}
	return nil
}
