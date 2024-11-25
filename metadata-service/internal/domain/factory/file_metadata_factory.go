package factory

import (
	"github.com/Zando74/GopherFS/metadata-service/internal/domain/entity"
)

type FileMetadataInput struct {
	FileID   string
	FileName string
	FileSize int64
	Chunks   []entity.ChunkLocations
}

type FileMetadataFactory struct{}

func (fmf FileMetadataFactory) CreateFileMetadata(input FileMetadataInput) (*entity.FileMetadata, error) {
	return &entity.FileMetadata{
		FileID:   input.FileID,
		FileName: input.FileName,
		FileSize: input.FileSize,
		Chunks:   input.Chunks,
	}, nil
}
