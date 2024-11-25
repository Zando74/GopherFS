package controller

import (
	"github.com/Zando74/GopherFS/metadata-service/internal/application/adapter/grpc"
	"github.com/Zando74/GopherFS/metadata-service/internal/domain/usecase"
)

type GetFileMetadataServer struct {
	grpc.UnimplementedFileMetadataServiceServer
	getFileMetadataUseCase usecase.GetFileMetadataUseCase
}

func (fs *GetFileMetadataServer) GetFileMetadata(stream grpc.FileMetadataService_GetFileMetadataServer) error {

	req, err := stream.Recv()
	if err != nil {
		return err
	}

	fileMetadata, err := fs.getFileMetadataUseCase.Execute(req.GetFileId())

	if err != nil {
		return err
	}

	chunkLocations := make([]*grpc.ChunkLocations, len(fileMetadata.Chunks))
	for i, chunk := range fileMetadata.Chunks {
		chunkLocations[i] = &grpc.ChunkLocations{
			ChunkId:        chunk.ChunkID,
			SequenceNumber: uint32(chunk.SequenceNumber),
			PrimaryLocation: &grpc.Location{
				StorageId: chunk.PrimaryLocation.StorageID,
			},
			SecondaryLocations: make([]*grpc.Location, len(chunk.SecondaryLocations)),
		}
	}

	return stream.SendAndClose(&grpc.GetFileMetadataResponse{
		FileId:         fileMetadata.FileID,
		FileName:       fileMetadata.FileName,
		FileSize:       uint32(fileMetadata.FileSize),
		ChunkLocations: chunkLocations,
	})
}
