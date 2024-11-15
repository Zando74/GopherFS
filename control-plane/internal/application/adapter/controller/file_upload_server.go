package controller

import (
	"io"
	"sync"

	"github.com/Zando74/GopherFS/control-plane/config"
	"github.com/Zando74/GopherFS/control-plane/internal/application/adapter/grpc"
	"github.com/Zando74/GopherFS/control-plane/internal/application/adapter/repository"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/usecase"
	"github.com/Zando74/GopherFS/control-plane/logger"
)

var (
	fileChunkRepositoryImpl = &repository.FileChunkRepository{}
)

type FileUploaderver struct {
	grpc.UnimplementedFileUploadServiceServer
	splitFileStream usecase.SplitFileStreamUseCase
	wg              sync.WaitGroup
}

func NewFileUploadServer() *FileUploaderver {
	return &FileUploaderver{
		splitFileStream: usecase.SplitFileStreamUseCase{
			FileChunkRepository: fileChunkRepositoryImpl,
		},
	}
}

func (fs *FileUploaderver) executeSplitFileStreamUseCase(filename string, fileBatchID uint32, batch []byte) {
	fs.wg.Add(1)
	go func(filename string, fileBatchID uint32, batch []byte) {
		defer fs.wg.Done()
		fs.splitFileStream.SplitFileBatch(filename, fileBatchID, batch)
	}(filename, fileBatchID, batch)
}

func (fs *FileUploaderver) UploadFile(stream grpc.FileUploadService_UploadFileServer) error {
	logger := logger.LoggerSingleton.GetInstance()
	config := config.ConfigSingleton.GetInstance()
	var fileSize uint32 = 0
	var batchCpt uint32 = 0
	filename := ""
	var buffer []byte
	for {
		req, err := stream.Recv()

		if filename == "" {
			filename = req.GetFileName()
		}

		if err == io.EOF {
			if len(buffer) > 0 {
				fs.executeSplitFileStreamUseCase(filename, batchCpt, buffer)
				fileSize += uint32(len(buffer))
			}
			break
		}

		if err != nil {
			logger.Error(err)
			// should think about a rollback strategy
			break
		}

		batch := req.GetBatch()
		buffer = append(buffer, batch...)

		if len(buffer) < int(config.FileStorage.Chunk_size) {
			continue
		}

		fs.executeSplitFileStreamUseCase(filename, batchCpt, buffer)

		fileSize += uint32(len(buffer))
		batchCpt++
		buffer = []byte{}
	}
	fs.wg.Wait()

	return stream.SendAndClose(&grpc.FileUploadResponse{FileName: filename, Size: fileSize, Message: "OK"})

	// should call a use case to start Saving Saga
}
