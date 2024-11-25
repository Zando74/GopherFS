package usecase

import (
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
)

type MakeBatchFromChunkUseCase struct{}

func (m *MakeBatchFromChunkUseCase) Execute(
	fileID string,
	chunks entity.FileChunk,
	callback func(filename string, batch []byte) error,
) error {

	const maxBatchSize = 1 * 1024 * 1024 // 1MB
	var errorThrown error

	batch := make([]byte, 0, maxBatchSize)

	chunkData := chunks.ChunkData
	for len(chunkData) > 0 {
		remainingSpace := maxBatchSize - len(batch)
		if remainingSpace == 0 {
			err := callback(fileID, batch)
			if err != nil {
				errorThrown = err
				return errorThrown
			}
			batch = make([]byte, 0, maxBatchSize)
			remainingSpace = maxBatchSize
		}

		if len(chunkData) <= remainingSpace {
			batch = append(batch, chunkData...)
			chunkData = nil
		} else {
			batch = append(batch, chunkData[:remainingSpace]...)
			chunkData = chunkData[remainingSpace:]
		}
	}

	if len(batch) > 0 {
		err := callback(fileID, batch)
		if err != nil {
			errorThrown = err
		}
	}

	return errorThrown
}
