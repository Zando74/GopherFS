package usecase

import (
	"github.com/Zando74/GopherFS/metadata-service/internal/domain/entity"
	"github.com/Zando74/GopherFS/metadata-service/internal/domain/repository"
)

type GetFileMetadataUseCase struct {
	FileMetadataRepository repository.FileMetadataRepository
}

func (g *GetFileMetadataUseCase) Execute(fileID string) (*entity.FileMetadata, error) {
	fileMetadata, err := g.FileMetadataRepository.GetFileMetadata(fileID)
	if err != nil {
		return nil, err
	}

	return fileMetadata, nil
}
