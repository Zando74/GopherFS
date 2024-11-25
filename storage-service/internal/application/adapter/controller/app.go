package controller

import (
	"net"

	"github.com/Zando74/GopherFS/storage-service/config"
	pb "github.com/Zando74/GopherFS/storage-service/internal/application/adapter/grpc"
	"github.com/Zando74/GopherFS/storage-service/internal/application/adapter/repository"
	"github.com/Zando74/GopherFS/storage-service/internal/domain/usecase"
	"github.com/Zando74/GopherFS/storage-service/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	cfg                     = config.ConfigSingleton.GetInstance()
	log                     = logger.LoggerSingleton.GetInstance()
	fileChunkRepositoryImpl = repository.NewStoreFileChunkImpl(cfg.BaseDirectory)
	onFileChunkSave         = NewOnFileChunkSave()
	onFileDelete            = NewOnFileDelete()
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

	pb.RegisterFileChunkDownloadServiceServer(c.grpcServer, &FileChunkDownloadServer{
		retrieveChunkForFileUseCase: usecase.RetrieveChunkForFileUseCase{
			FileChunkRepository: fileChunkRepositoryImpl,
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

	go onFileChunkSave.Listen()
	go onFileDelete.Listen()

	<-c.stopSignal
	log.Debug("Shutting down gRPC server")
	c.grpcServer.Stop()
}
