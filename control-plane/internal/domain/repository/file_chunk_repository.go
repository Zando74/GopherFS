package repository

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
)

type FileChunkRepository interface {
	SaveFileChunk(chunk *entity.FileChunk, onSuccess func(fileChunkLocation entity.ChunkLocations)) error
	DeleteAllFilesChunks(fileID string) error
}
