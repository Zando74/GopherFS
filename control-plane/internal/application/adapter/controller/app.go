package controller

import (
	"net"

	"github.com/Zando74/GopherFS/control-plane/config"
	"github.com/Zando74/GopherFS/control-plane/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/Zando74/GopherFS/control-plane/internal/application/adapter/grpc"
	"github.com/Zando74/GopherFS/control-plane/internal/application/adapter/repository"
)

var (
	cfg                           = config.ConfigSingleton.GetInstance()
	fileChunkRepositoryImpl       = &repository.FileChunkRepository{}
	fileMetadataRepositoryImpl    = &repository.FileMetadataRepository{}
	fileReplicationRequestImpl    = &repository.FileReplicationRequestRepository{}
	sagaInformationRepositoryImpl = repository.NewSagaInformationRepository()
	log                           = logger.LoggerSingleton.GetInstance()
)

type Controller struct {
	IsLeader   bool
	grpcServer *grpc.Server
	stopSignal chan bool
}

func (c *Controller) Shutdown() {
	c.stopSignal <- true
}

func (c *Controller) OnLeaderElection() {
	c.IsLeader = true
	logger.LoggerSingleton.GetInstance().Info("Check pending saga")
	ProcessPendingSagaAtInitController := &ProcessPendingSagaAtInitController{}
	ProcessPendingSagaAtInitController.Run()
}

func (c *Controller) Run() {
	log.Info(logger.RunningApplicationMessage)

	c.grpcServer = grpc.NewServer()
	reflection.Register(c.grpcServer)

	pb.RegisterFileUploadServiceServer(c.grpcServer, &FileUploaderver{})

	listen, err := net.Listen("tcp", cfg.GRPC.Port)
	if err != nil {
		log.Fatal(err)
	}

	go func(g *grpc.Server) {
		if err := g.Serve(listen); err != nil {
			log.Fatal(err)
		}
	}(c.grpcServer)

	<-c.stopSignal
	log.Debug("Shutting down gRPC server")
	c.grpcServer.Stop()
}
