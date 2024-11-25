package repository

import (
	"github.com/Zando74/GopherFS/metadata-service/internal/domain/entity"
)

type FileMetadataRepository interface {
	GetFileMetadata(fileID string) (*entity.FileMetadata, error)
	SaveFileMetadata(metadata *entity.FileMetadata, onSuccess func()) error
	DeleteFileMetadata(fileID string) error
}
