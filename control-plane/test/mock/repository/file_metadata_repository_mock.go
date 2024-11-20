package mock_repository

import (
	"math/rand"
	"time"

	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
)

type FileMetadataRepository struct {
	metadataStorage []*entity.FileMetadata
}

func (fcr *FileMetadataRepository) SaveFileMetadata(metadata *entity.FileMetadata, onSuccess func()) error {
	fcr.metadataStorage = append(fcr.metadataStorage, metadata)
	go func() {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		onSuccess()
	}()

	return nil
}

func (fcr *FileMetadataRepository) DeleteFileMetadata(fileID string) error {
	for i, metadata := range fcr.metadataStorage {
		if metadata.FileID == fileID {
			fcr.metadataStorage = append(fcr.metadataStorage[:i], fcr.metadataStorage[i+1:]...)
			break
		}
	}
	return nil
}
