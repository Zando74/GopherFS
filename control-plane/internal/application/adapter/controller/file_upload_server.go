package controller

import (
	"io"
	"sync"

	"github.com/Zando74/GopherFS/control-plane/internal/application/adapter/grpc"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/saga/coordinator"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/usecase"
	"github.com/google/uuid"
)

type FileUploaderver struct {
	grpc.UnimplementedFileUploadServiceServer
	splitBatchToChunkUseCase usecase.SplitBatchToChunkUseCase
	startFileStreamSaga      usecase.StartFileUploadSagaUseCase
	wg                       sync.WaitGroup
}

func (fs *FileUploaderver) executeSplitFileStreamUseCase(
	filename string,
	fileID string,
	fileBatchID uint32,
	batch []byte,
) {
	fs.wg.Add(1)
	go func(filename string, fileID string, fileBatchID uint32, batch []byte) {
		defer fs.wg.Done()
		fs.splitBatchToChunkUseCase.Execute(filename, fileID, fileBatchID, batch)
	}(filename, fileID, fileBatchID, batch)
}

func (fs *FileUploaderver) UploadFile(stream grpc.FileUploadService_UploadFileServer) error {

	var fileSize uint32 = 0
	var batchCpt uint32 = 0
	filename := ""
	fileID := uuid.New().String()
	var buffer []byte
	var uploadSagaCoordinator *coordinator.UploadSagaCoordinator
	for {
		req, err := stream.Recv()

		if filename == "" {
			filename = req.GetFileName()
			uploadSagaCoordinator = coordinator.NewUploadSagaCoordinator(
				fileChunkRepositoryImpl,
				fileMetadataRepositoryImpl,
				fileReplicationRequestImpl,
				sagaInformationRepositoryImpl,
				fileID, // generate a new UUID for the ID
				filename,
			)
			fs.splitBatchToChunkUseCase = usecase.SplitBatchToChunkUseCase{
				UploadSagaCoordinator: uploadSagaCoordinator,
			}
			fs.startFileStreamSaga = usecase.StartFileUploadSagaUseCase{
				UploadSagaCoordinator: uploadSagaCoordinator,
			}
			go func() { uploadSagaCoordinator.StartSaga() }()

		}

		if err == io.EOF {
			if len(buffer) > 0 {
				fs.executeSplitFileStreamUseCase(filename, fileID, batchCpt, buffer)
				fileSize += uint32(len(buffer))
			}
			break
		}

		if err != nil {
			log.Error(err)
			uploadSagaCoordinator.RollbackSaga()
			break
		}

		batch := req.GetBatch()
		buffer = append(buffer, batch...)

		if len(buffer) < int(cfg.FileStorage.Chunk_size) {
			continue
		}

		fs.executeSplitFileStreamUseCase(filename, fileID, batchCpt, buffer)

		fileSize += uint32(len(buffer))
		batchCpt++
		buffer = []byte{}
	}
	fs.wg.Wait()

	uploadSagaCoordinator.StartWaitingForConfirmations()

	return stream.SendAndClose(&grpc.FileUploadResponse{FileName: filename, Size: fileSize, Message: "OK"})

	// should call a use case to start Saving Saga
}
