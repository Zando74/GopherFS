package usecase

import (
	"github.com/Zando74/GopherFS/storage-service/internal/domain/entity"
	"github.com/Zando74/GopherFS/storage-service/internal/domain/factory"
	"github.com/Zando74/GopherFS/storage-service/internal/domain/repository"
)

type StoreFileChunkUseCase struct {
	FileChunkRepository repository.FileChunkRepository
	FileChunkFactory    factory.FileChunkFactory
}

func (s *StoreFileChunkUseCase) Execute(
	fileID string,
	sequenceNumber uint32,
	batch []byte,
	onSuccess func(fileChunkLocation entity.ChunkLocations),
) error {

	filechunk, err := s.FileChunkFactory.CreateFileChunk(factory.FileChunkInput{
		FileID:         fileID,
		SequenceNumber: sequenceNumber,
		StableContext:  fileID,
		ChunkData:      batch,
	})

	if err != nil {
		return err
	}

	location, err := s.FileChunkRepository.SaveFileChunk(filechunk)

	if err != nil {
		return err
	}
	onSuccess(*location)
	return nil

}
