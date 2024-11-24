package test

import (
	"testing"

	"github.com/Zando74/GopherFS/control-plane/internal/domain/usecase"
)

func TestSplitFileStreamUseCase(t *testing.T) {

	splitBatchToChunkUseCase, uploadSagaCoordinator, fileChunkTestRepository := GenerateFileUploadSagaUseCase("test", "test")

	// Generate a 60MB byte array
	data := make([]byte, 60*1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Start Saga to split the file into chunks
	startFileStreamSagaUC := usecase.StartFileUploadSagaUseCase{
		UploadSagaCoordinator: uploadSagaCoordinator,
	}
	startFileStreamSagaUC.Execute()

	// Read the data in 1MB batches
	for i := 0; i < 10; i++ {
		start := i * 1024 * 1024
		end := start + 1024*1024
		batch := data[start:end]

		// Process the batch
		err := splitBatchToChunkUseCase.Execute("test", "test", uint32(i), batch)

		if err != nil {
			t.Errorf("Error processing batch %d: %v", i, err)
		}

		// Verify that the batch was processed correctly
		fileChunks := fileChunkTestRepository.GetFiles()
		if len(fileChunks) != i+1 {
			t.Errorf("Expected %d chunks, got %d", i+1, len(fileChunks))
		}

		// Verify that the data in the chunks is correct
		for j, chunk := range fileChunks {
			start := j * 1024 * 1024
			end := start + 1024*1024
			expected := data[start:end]
			if string(chunk.ChunkData) != string(expected) {
				t.Errorf("Chunk %d data mismatch", j)
			}
		}
	}

}

func TestSplitFileStreamUseCase_EmptyBatch(t *testing.T) {

	splitBatchToChunkUseCase, uploadSagaCoordinator, fileChunkTestRepository := GenerateFileUploadSagaUseCase("test_empty", "test_empty")

	// Generate an empty byte array
	data := make([]byte, 0)

	// Start Saga to split the file into chunks
	startFileStreamSagaUC := usecase.StartFileUploadSagaUseCase{
		UploadSagaCoordinator: uploadSagaCoordinator,
	}
	startFileStreamSagaUC.Execute()

	// Process the empty batch
	err := splitBatchToChunkUseCase.Execute("test_empty", "test_empty", 0, data)

	if err != nil {
		t.Errorf("Error processing empty batch: %v", err)
	}

	// Verify that no chunks were created
	fileChunks := fileChunkTestRepository.GetFiles()
	if len(fileChunks) != 0 {
		t.Errorf("Expected 0 chunks, got %d", len(fileChunks))
	}
}
