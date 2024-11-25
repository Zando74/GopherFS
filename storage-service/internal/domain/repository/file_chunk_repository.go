package repository

import (
	"github.com/Zando74/GopherFS/storage-service/internal/domain/entity"
)

type FileChunkRepository interface {
	SaveFileChunk(chunk *entity.FileChunk) (*entity.ChunkLocations, error)
	RetrieveChunkForFile(fileID string, sequenceNumber uint32) (*entity.FileChunk, error)
	DeleteAllFilesChunks(fileID string) error
}
