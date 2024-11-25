package usecase

import (
	"github.com/Zando74/GopherFS/storage-service/internal/domain/entity"
	"github.com/Zando74/GopherFS/storage-service/internal/domain/repository"
)

type RetrieveChunkForFileUseCase struct {
	FileChunkRepository repository.FileChunkRepository
}

func (r *RetrieveChunkForFileUseCase) Execute(
	fileID string,
	sequenceNumber uint32,
) (*entity.FileChunk, error) {
	fileChunk, err := r.FileChunkRepository.RetrieveChunkForFile(fileID, sequenceNumber)
	if err != nil {
		return nil, err
	}
	return fileChunk, nil
}
