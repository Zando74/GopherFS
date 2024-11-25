package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Zando74/GopherFS/storage-service/config"
	"github.com/Zando74/GopherFS/storage-service/internal/domain/entity"
	"github.com/Zando74/GopherFS/storage-service/internal/domain/value_object"
	"github.com/Zando74/GopherFS/storage-service/logger"
)

type StoreFileChunkImpl struct {
	BaseDirectory string
	cfg           *config.Config
}

func NewStoreFileChunkImpl(baseDirectory string) *StoreFileChunkImpl {
	cfg := config.ConfigSingleton.GetInstance()
	return &StoreFileChunkImpl{
		BaseDirectory: baseDirectory,
		cfg:           cfg,
	}
}

func (s *StoreFileChunkImpl) SaveFileChunk(chunk *entity.FileChunk) (*entity.ChunkLocations, error) {
	fileDir := filepath.Join(s.BaseDirectory, chunk.FileID)
	err := os.MkdirAll(fileDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	filePath := filepath.Join(fileDir, fmt.Sprintf("%d.chunk", chunk.SequenceNumber))
	err = os.WriteFile(filePath, chunk.ChunkData, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to write chunk to file: %w", err)
	}

	logger.LoggerSingleton.GetInstance().Info(fmt.Sprintf("chunk saved: %s", filePath))

	return &entity.ChunkLocations{
		ChunkID:        string(chunk.ChunkID),
		SequenceNumber: chunk.SequenceNumber,
		PrimaryLocation: entity.Location{
			StorageID: s.cfg.Storage.ServiceName,
		},
		SecondaryLocations: []entity.Location{},
	}, nil
}

func (s *StoreFileChunkImpl) RetrieveChunkForFile(fileID string, sequenceNumber uint32) (*entity.FileChunk, error) {
	chunkFileName := fmt.Sprintf("%d.chunk", sequenceNumber)
	filePath := filepath.Join(s.BaseDirectory, fileID, chunkFileName)

	chunkData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk file: %w", err)
	}

	return &entity.FileChunk{
		FileID:         fileID,
		SequenceNumber: sequenceNumber,
		ChunkID:        value_object.NewChunkHash(chunkData, fileID),
		ChunkData:      chunkData,
	}, nil
}

func (s *StoreFileChunkImpl) DeleteAllFilesChunks(fileID string) error {
	chunkDir := filepath.Join(s.cfg.Storage.BaseDirectory, fileID)
	files, err := os.ReadDir(chunkDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".chunk") {
			err = os.Remove(filepath.Join(chunkDir, file.Name()))
			if err != nil {
				return fmt.Errorf("failed to delete chunk file: %w", err)
			}
		}
	}

	err = os.Remove(chunkDir)
	if err != nil {
		return fmt.Errorf("failed to delete directory: %w", err)
	}

	logger.LoggerSingleton.GetInstance().Info(fmt.Sprintf("chunk deleted: %s", chunkDir))

	return nil
}
