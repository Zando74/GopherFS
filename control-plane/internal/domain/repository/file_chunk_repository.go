package repository

import "github.com/Zando74/GopherFS/control-plane/internal/domain/entity"

type FileChunkRepository interface {
	SaveFileChunk(chunk *entity.FileChunk) error
}
