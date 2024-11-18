package repository

import (
	"math/rand"
	"time"

	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
)

type FileMetadataRepository struct{}

func (fcr FileMetadataRepository) SaveFileMetadata(metadata *entity.FileMetadata, onSuccess func()) error {

	go func() {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		onSuccess()
	}()

	return nil
}

func (fcr FileMetadataRepository) DeleteFileMetadata(fileID string) error {
	return nil
}
