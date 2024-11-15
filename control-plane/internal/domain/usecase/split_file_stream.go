package usecase

import (
	"fmt"
	"sync"

	"github.com/Zando74/GopherFS/control-plane/config"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/factory"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/repository"
	"github.com/Zando74/GopherFS/control-plane/logger"
)

type SplitFileStreamUseCase struct {
	FileChunkFactory    factory.FileChunkFactory
	FileChunkRepository repository.FileChunkRepository
}

func (s *SplitFileStreamUseCase) SplitFileBatch(filename string, fileBatchID uint32, batch []byte) error {

	config := config.ConfigSingleton.GetInstance()
	chunkSize := config.FileStorage.Chunk_size
	chunkCpt := 0
	var errorThrown error

	wg := &sync.WaitGroup{}
	for i := uint32(0); i < uint32(len(batch)); i += chunkSize {
		if i+chunkSize > uint32(len(batch)) {
			chunkSize = uint32(len(batch)) - i
		}

		wg.Add(1)
		go func(chunkID int) {
			defer wg.Done()
			chunk := batch[i : i+chunkSize]
			// send chunk to the next step
			filechunk, err := s.FileChunkFactory.CreateFileChunk(factory.FileChunkInput{
				Filename:       filename,
				SequenceNumber: fmt.Sprintf("%d-%d", fileBatchID, chunkID),
				StableContext:  fmt.Sprintf("%d-%d", fileBatchID, chunkID),
				ChunkData:      chunk,
			})

			if err != nil {
				errorThrown = err
				return
			}
			logger := logger.LoggerSingleton.GetInstance()
			logger.Debug("process batch number : %d , chunk hash : %s, sequence number : %d, chunk length: %d", fileBatchID, filechunk.ChunkID, filechunk.SequenceNumber, len(filechunk.ChunkData))

			s.FileChunkRepository.SaveFileChunk(filechunk)
		}(chunkCpt)
		chunkCpt++
	}
	wg.Wait()

	return errorThrown
}
