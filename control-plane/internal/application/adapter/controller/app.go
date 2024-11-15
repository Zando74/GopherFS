package controller

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Zando74/GopherFS/control-plane/config"
	"github.com/Zando74/GopherFS/control-plane/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/Zando74/GopherFS/control-plane/internal/application/adapter/grpc"
)

func Run() {
	logger := logger.LoggerSingleton.GetInstance()
	config := config.ConfigSingleton.GetInstance()

	logger.Info("Running application")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g := grpc.NewServer()
	reflection.Register(g)

	// should register services here
	// ...
	pb.RegisterFileUploadServiceServer(g, NewFileUploadServer())

	listen, err := net.Listen("tcp", config.GRPC.Port)
	if err != nil {
		logger.Fatal(err)
	}

	interrupt := make(chan os.Signal, 1)
	shutdownSignals := []os.Signal{
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	}
	signal.Notify(interrupt, shutdownSignals...)
	go func(g *grpc.Server) {
		logger.Info("setGRPC - gRPC server started on " + config.GRPC.Port)
		if err := g.Serve(listen); err != nil {
			logger.Fatal(err)
		}
	}(g)
	select {
	case killSignal := <-interrupt:
		logger.Debug("Got ", killSignal)
	case <-ctx.Done():
	}

}
