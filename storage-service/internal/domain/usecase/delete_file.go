package usecase

import "github.com/Zando74/GopherFS/storage-service/internal/domain/repository"

type DeleteFileUseCase struct {
	FileChunkRepository repository.FileChunkRepository
}

func (d *DeleteFileUseCase) Execute(fileID string) error {
	err := d.FileChunkRepository.DeleteAllFilesChunks(fileID)
	return err
}
