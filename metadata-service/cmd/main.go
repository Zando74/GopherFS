package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Zando74/GopherFS/metadata-service/config"
	"github.com/Zando74/GopherFS/metadata-service/internal/application/adapter/controller"
	"github.com/Zando74/GopherFS/metadata-service/logger"
)

func main() {
	cfg := config.ConfigSingleton.GetInstance()
	log := logger.LoggerSingleton.GetInstance()
	controller := controller.Controller{}

	log.Info(logger.ConfigLoadedMessage, cfg)

	controller.Run()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	controller.Shutdown()
	os.Exit(0)
}
