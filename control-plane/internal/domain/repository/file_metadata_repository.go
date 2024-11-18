package repository

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
)

type FileMetadataRepository interface {
	SaveFileMetadata(metadata *entity.FileMetadata, onSuccess func()) error
	DeleteFileMetadata(fileID string) error
}
