package usecase

import "github.com/Zando74/GopherFS/metadata-service/internal/domain/repository"

type DeleteFileMetadataUseCase struct {
	FileMetadataRepository repository.FileMetadataRepository
}

func (d *DeleteFileMetadataUseCase) Execute(fileID string) error {
	return d.FileMetadataRepository.DeleteFileMetadata(fileID)
}
