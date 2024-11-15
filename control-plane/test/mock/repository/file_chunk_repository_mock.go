package mock_repository

import "github.com/Zando74/GopherFS/control-plane/internal/domain/entity"

type FileChunkRepositoryMock struct {
	files []*entity.FileChunk
}

func (fcrm *FileChunkRepositoryMock) SaveFileChunk(chunk *entity.FileChunk) error {
	fcrm.files = append(fcrm.files, chunk)
	return nil
}

func (fcrm *FileChunkRepositoryMock) GetFiles() []*entity.FileChunk {
	return fcrm.files
}
