package test

import (
	"sync"
	"testing"

	"github.com/Zando74/GopherFS/control-plane/internal/domain/saga/coordinator"
	saga_entity "github.com/Zando74/GopherFS/control-plane/internal/domain/saga/entity"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/usecase"
	mock_repository "github.com/Zando74/GopherFS/control-plane/test/mock/repository"
)

func TestStartFileUploadSaga(t *testing.T) {

	var fileChunkTestRepository = &mock_repository.FileChunkRepositoryMock{}
	var fileMetadataRepository = &mock_repository.FileMetadataRepository{}
	var fileReplicationRequestRepository = &mock_repository.FileReplicationRequestRepository{}
	var sagaInformationRepository = &mock_repository.SagaInformationRepositoryImpl{}

	var uploadSagaCoordinator = coordinator.NewUploadSagaCoordinator(
		fileChunkTestRepository,
		fileMetadataRepository,
		fileReplicationRequestRepository,
		sagaInformationRepository,
		"test1",
		"test1",
	)

	var splitBatchToChunkUseCase = &usecase.SplitBatchToChunkUseCase{
		UploadSagaCoordinator: uploadSagaCoordinator,
	}

	// Generate a 60MB byte array
	data := make([]byte, 60*1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Start Saga to split the file into chunks
	startFileStreamSagaUC := usecase.StartFileUploadSagaUseCase{
		UploadSagaCoordinator: uploadSagaCoordinator,
	}
	wg := &sync.WaitGroup{}
	go func() {
		defer wg.Done()
		startFileStreamSagaUC.ExecuteSync()
	}()
	wg.Add(1)

	// Read the data in 1MB batches
	for i := 0; i < 10; i++ {
		start := i * 1024 * 1024
		end := start + 1024*1024
		batch := data[start:end]

		// Process the batch
		err := splitBatchToChunkUseCase.Execute("test1", uint32(i), batch)

		if err != nil {
			t.Errorf("Error processing batch %d: %v", i, err)
		}

		// Verify that the batch was processed correctly
		fileChunks := fileChunkTestRepository.GetFiles()
		if len(fileChunks) != i+1 {
			t.Errorf("Expected %d chunks, got %d", i+1, len(fileChunks))
		}
	}
	uploadSagaCoordinator.StartWaitingForConfirmations()
	wg.Wait()

	// verify that each steps are executed correctly
	pendingSteps, _ := sagaInformationRepository.GetPendingSagaInformation()
	if len(pendingSteps) != 0 {
		t.Errorf("Expected 0 pending steps, got %d", len(pendingSteps))
	}

	failedSteps, _ := sagaInformationRepository.GetFailedSagaInformation()
	if len(failedSteps) != 0 {
		t.Errorf("Expected 0 done steps, got %d", len(failedSteps))
	}
}

func TestAbortedUploadSagaRecovery(t *testing.T) {
	var fileChunkTestRepository = &mock_repository.FileChunkRepositoryMock{}
	var fileMetadataRepository = &mock_repository.FileMetadataRepository{}
	var fileReplicationRequestRepository = &mock_repository.FileReplicationRequestRepository{}
	var sagaInformationRepository = &mock_repository.SagaInformationRepositoryImpl{}

	var uploadSagaCoordinator = coordinator.NewUploadSagaCoordinator(
		fileChunkTestRepository,
		fileMetadataRepository,
		fileReplicationRequestRepository,
		sagaInformationRepository,
		"test2",
		"test2",
	)

	var splitBatchToChunkUseCase = &usecase.SplitBatchToChunkUseCase{
		UploadSagaCoordinator: uploadSagaCoordinator,
	}

	// Generate a 60MB byte array
	data := make([]byte, 60*1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Start Saga to split the file into chunks
	startFileStreamSagaUC := usecase.StartFileUploadSagaUseCase{
		UploadSagaCoordinator: uploadSagaCoordinator,
	}
	wg := &sync.WaitGroup{}
	go func() {
		defer wg.Done()
		startFileStreamSagaUC.ExecuteSync()
	}()
	wg.Add(1)

	// Read the data in 1MB batches
	for i := 0; i < 10; i++ {
		start := i * 1024 * 1024
		end := start + 1024*1024
		batch := data[start:end]

		// Process the batch
		splitBatchToChunkUseCase.Execute("test2", uint32(i), batch)
		// Verify that the batch was processed correctly
	}
	uploadSagaCoordinator.StartWaitingForConfirmations()
	// Abort the Saga immediately during process
	uploadSagaCoordinator.RollbackSaga()

	wg.Wait()

	fileChunks := fileChunkTestRepository.GetFiles()
	if len(fileChunks) != 0 {
		t.Errorf("Expected %d chunks, got %d", 0, len(fileChunks))
	}

	// Verify that the saga was aborted
	pendingSteps, _ := sagaInformationRepository.GetPendingSagaInformation()
	if len(pendingSteps) != 0 {
		t.Errorf("Expected 0 pending steps, got %d %v", len(pendingSteps), pendingSteps)
	}

	// Log the failed saga steps
	failedSteps, _ := sagaInformationRepository.GetFailedSagaInformation()
	if len(failedSteps) > 0 {
		t.Logf("Failed saga steps: %v", failedSteps)
	} else {
		t.Log("No failed saga steps")
	}
}

func TestAbortedFileMetadataStepRecovery(t *testing.T) {
	var fileChunkTestRepository = &mock_repository.FileChunkRepositoryMock{}
	var fileMetadataRepository = &mock_repository.FileMetadataRepository{}
	var fileReplicationRequestRepository = &mock_repository.FileReplicationRequestRepository{}
	var sagaInformationRepository = &mock_repository.SagaInformationRepositoryImpl{}

	var uploadSagaCoordinator = coordinator.NewUploadSagaCoordinator(
		fileChunkTestRepository,
		fileMetadataRepository,
		fileReplicationRequestRepository,
		sagaInformationRepository,
		"test3",
		"test3",
	)

	var splitBatchToChunkUseCase = &usecase.SplitBatchToChunkUseCase{
		UploadSagaCoordinator: uploadSagaCoordinator,
	}

	// Generate a 60MB byte array
	data := make([]byte, 60*1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Start Saga to split the file into chunks
	startFileStreamSagaUC := usecase.StartFileUploadSagaUseCase{
		UploadSagaCoordinator: uploadSagaCoordinator,
	}
	wg := &sync.WaitGroup{}
	go func() {
		defer wg.Done()
		startFileStreamSagaUC.ExecuteSync()
	}()
	wg.Add(1)

	// Read the data in 1MB batches
	for i := 0; i < 10; i++ {
		start := i * 1024 * 1024
		end := start + 1024*1024
		batch := data[start:end]

		// Process the batch
		splitBatchToChunkUseCase.Execute("test3", uint32(i), batch)
		// Verify that the batch was processed correctly
	}
	uploadSagaCoordinator.StartWaitingForConfirmations()
	// Abort the Saga immediately during process
	uploadSagaCoordinator.Saga.Steps[1].Fail()
	wg.Wait()

	fileChunks := fileChunkTestRepository.GetFiles()
	if len(fileChunks) != 0 {
		t.Errorf("Expected %d chunks, got %d", 0, len(fileChunks))
	}

	// Verify that the saga was aborted
	pendingSteps, _ := sagaInformationRepository.GetPendingSagaInformation()
	if len(pendingSteps) != 0 {
		t.Errorf("Expected 0 pending steps, got %d %v", len(pendingSteps), pendingSteps)
	}

	// Log the failed saga steps
	failedSteps, _ := sagaInformationRepository.GetFailedSagaInformation()
	if len(failedSteps) == 1 && failedSteps[0].StepName == saga_entity.ReplicateFile {
		t.Logf("Failed saga steps: %v", failedSteps)
	} else {
		t.Log("No failed saga steps")
	}
}

func TestAbortedFileReplicationStepRecovery(t *testing.T) {
	splitBatchToChunkUseCase, uploadSagaCoordinator, fileChunkTestRepository := GenerateFileUploadSagaUseCase("test4", "test4")

	// Generate a 60MB byte array
	data := make([]byte, 60*1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Start Saga to split the file into chunks
	startFileStreamSagaUC := usecase.StartFileUploadSagaUseCase{
		UploadSagaCoordinator: uploadSagaCoordinator,
	}
	wg := &sync.WaitGroup{}
	go func() {
		defer wg.Done()
		startFileStreamSagaUC.ExecuteSync()
	}()
	wg.Add(1)

	// Read the data in 1MB batches
	for i := 0; i < 10; i++ {
		start := i * 1024 * 1024
		end := start + 1024*1024
		batch := data[start:end]

		// Process the batch
		splitBatchToChunkUseCase.Execute("test4", uint32(i), batch)
		// Verify that the batch was processed correctly
	}
	uploadSagaCoordinator.StartWaitingForConfirmations()
	// Abort the Saga immediately during process
	uploadSagaCoordinator.Saga.Steps[2].Fail()
	wg.Wait()

	fileChunks := fileChunkTestRepository.GetFiles()
	if len(fileChunks) != 0 {
		t.Errorf("Expected %d chunks, got %d", 0, len(fileChunks))
	}

	// Verify that the saga was aborted
	pendingSteps, _ := uploadSagaCoordinator.FileChunkSaveConfirmationStep.SagaInformationLogger.GetPendingSagaInformation()
	if len(pendingSteps) != 0 {
		t.Errorf("Expected 0 pending steps, got %d %v", len(pendingSteps), pendingSteps)
	}

	// Log the failed saga steps
	failedSteps, _ := uploadSagaCoordinator.FileChunkSaveConfirmationStep.SagaInformationLogger.GetFailedSagaInformation()
	if len(failedSteps) == 1 && failedSteps[0].StepName == saga_entity.ReplicateFile {
		t.Logf("Failed saga steps: %v", failedSteps)
	} else {
		t.Error("No failed saga steps")
	}
}
