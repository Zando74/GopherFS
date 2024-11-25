package controller

import (
	"net"

	"github.com/Zando74/GopherFS/metadata-service/config"
	pb "github.com/Zando74/GopherFS/metadata-service/internal/application/adapter/grpc"
	"github.com/Zando74/GopherFS/metadata-service/internal/application/adapter/repository"
	"github.com/Zando74/GopherFS/metadata-service/internal/domain/usecase"
	"github.com/Zando74/GopherFS/metadata-service/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	cfg                        = config.ConfigSingleton.GetInstance()
	log                        = logger.LoggerSingleton.GetInstance()
	fileMetadataRepositoryImpl = repository.NewEtcdFileMetadataRepository()
	onFileMetadataSave         = NewOnFileMetadataSave()
	onFileMetadataDelete       = NewOnFileMetadataDelete()
)

type Controller struct {
	stopSignal chan bool
	grpcServer *grpc.Server
}

func (c *Controller) Shutdown() {
	c.stopSignal <- true
}

func (c *Controller) Run() {
	log.Info(logger.RunningApplicationMessage)

	c.grpcServer = grpc.NewServer()
	reflection.Register(c.grpcServer)

	pb.RegisterFileMetadataServiceServer(c.grpcServer, &GetFileMetadataServer{
		getFileMetadataUseCase: usecase.GetFileMetadataUseCase{
			FileMetadataRepository: fileMetadataRepositoryImpl,
		},
	})

	listen, err := net.Listen("tcp", cfg.GRPC.Port)
	if err != nil {
		log.Fatal(err)
	}

	go func(g *grpc.Server) {
		if err := g.Serve(listen); err != nil {
			log.Fatal(err)
		}
	}(c.grpcServer)

	go onFileMetadataSave.Listen()
	go onFileMetadataDelete.Listen()

	<-c.stopSignal
	log.Debug("Shutting down gRPC server")
	c.grpcServer.Stop()
}
