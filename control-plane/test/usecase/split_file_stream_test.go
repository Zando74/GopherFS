package test

import (
	"testing"

	"github.com/Zando74/GopherFS/control-plane/internal/domain/usecase"
	mock_repository "github.com/Zando74/GopherFS/control-plane/test/mock/repository"
)

var fileChunkTestRepository = &mock_repository.FileChunkRepositoryMock{}

var splitFileStreamUseCase = &usecase.SplitFileStreamUseCase{
	FileChunkRepository: fileChunkTestRepository,
}

func TestSplitFileStreamUseCase(t *testing.T) {

	// Generate a 10MB byte array
	data := make([]byte, 10*1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Read the data in 1MB batches
	for i := 0; i < 10; i++ {
		start := i * 1024 * 1024
		end := start + 1024*1024
		batch := data[start:end]

		// Process the batch (this is just an example, replace with actual processing code)
		t.Logf("Processing batch %d: %v", i, batch[:10]) // Log first 10 bytes of each batch
		err := splitFileStreamUseCase.SplitFileBatch("test", uint32(i), batch)

		if err != nil {
			t.Errorf("Error processing batch %d: %v", i, err)
		}

		// Verify that the batch was processed correctly
		fileChunks := fileChunkTestRepository.GetFiles()
		if len(fileChunks) != i+1 {
			t.Errorf("Expected %d chunks, got %d", i+1, len(fileChunks))
		}
	}

}
