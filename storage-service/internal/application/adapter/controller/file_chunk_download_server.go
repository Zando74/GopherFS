package controller

import (
	"github.com/Zando74/GopherFS/storage-service/internal/application/adapter/grpc"
	"github.com/Zando74/GopherFS/storage-service/internal/domain/usecase"
)

type FileChunkDownloadServer struct {
	grpc.UnimplementedFileChunkDownloadServiceServer
	retrieveChunkForFileUseCase usecase.RetrieveChunkForFileUseCase
}

func (fs *FileChunkDownloadServer) DownloadChunk(stream grpc.FileChunkDownloadService_DownloadChunkServer) error {

	req, err := stream.Recv()
	if err != nil {
		return err
	}

	fileId := req.GetFileId()
	sequenceNumber := req.GetSequenceNumber()

	fileChunk, err := fs.retrieveChunkForFileUseCase.Execute(fileId, sequenceNumber)

	if err != nil {
		return err
	}

	return stream.SendAndClose(&grpc.FileChunkDownloadResponse{
		FileId:         fileChunk.FileID,
		SequenceNumber: fileChunk.SequenceNumber,
		Chunk:          fileChunk.ChunkData,
	})

}
