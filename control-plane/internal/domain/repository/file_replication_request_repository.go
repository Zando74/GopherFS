package repository

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
)

type FileReplicationRequestRepository interface {
	SaveReplicationRequest(replicationRequest *entity.FileReplicationRequest, onSuccess func()) error
	DeleteReplicationRequest(replicationRequest *entity.FileReplicationRequest) error
}
