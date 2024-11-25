package usecase

import (
	"github.com/Zando74/GopherFS/metadata-service/internal/domain/entity"
	"github.com/Zando74/GopherFS/metadata-service/internal/domain/factory"
	"github.com/Zando74/GopherFS/metadata-service/internal/domain/repository"
)

type StoreFileMetadataUseCase struct {
	FileMetadataRepository repository.FileMetadataRepository
	FileMetadataFactory    factory.FileMetadataFactory
}

func (s *StoreFileMetadataUseCase) Execute(
	fileID string,
	fileName string,
	fileSize int64,
	chunks []entity.ChunkLocations,
	onSuccess func(),
) error {
	metadata, err := s.FileMetadataFactory.CreateFileMetadata(factory.FileMetadataInput{
		FileID:   fileID,
		FileName: fileName,
		FileSize: fileSize,
		Chunks:   chunks,
	})

	if err != nil {
		return err
	}

	err = s.FileMetadataRepository.SaveFileMetadata(metadata, onSuccess)

	if err != nil {
		return err
	}

	return nil
}
