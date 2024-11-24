package usecase

import (
	"fmt"
	"sync"

	"github.com/Zando74/GopherFS/control-plane/config"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/factory"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/repository"
)

type SplitBatchToChunkUseCase struct {
	FileChunkFactory      factory.FileChunkFactory
	UploadSagaCoordinator repository.SagaCoordinator
}

func (s *SplitBatchToChunkUseCase) Execute(
	filename string,
	fileID string,
	fileBatchID uint32,
	batch []byte,
) error {

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
		go func(chunkID int, index uint32, currentChunksize uint32) {
			defer wg.Done()

			chunk := batch[index : index+currentChunksize]

			filechunk, err := s.FileChunkFactory.CreateFileChunk(factory.FileChunkInput{
				Filename:       filename,
				SequenceNumber: fmt.Sprintf("%d-%d", fileBatchID, chunkID),
				StableContext:  fileID,
				ChunkData:      chunk,
			})

			if err != nil {
				s.UploadSagaCoordinator.RollbackSaga()
				errorThrown = err
				return
			}

			err = s.UploadSagaCoordinator.CommitFileChunk(filechunk)

			if err != nil {
				s.UploadSagaCoordinator.RollbackSaga()
				errorThrown = err
				return
			}

		}(chunkCpt, i, chunkSize)
		chunkCpt++

	}
	wg.Wait()

	return errorThrown
}
