package repository

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
)

type FileChunkRepository struct{}

func (fcr FileChunkRepository) SaveFileChunk(chunk *entity.FileChunk) error {
	return nil
}
