package main

import (
	"github.com/Zando74/GopherFS/control-plane/config"
	"github.com/Zando74/GopherFS/control-plane/internal/application/adapter/controller"
	"github.com/Zando74/GopherFS/control-plane/logger"
)

func main() {

	cfg := config.ConfigSingleton.GetInstance()
	logger := logger.LoggerSingleton.GetInstance()

	logger.Info("config: %s", cfg)

	controller.Run()

}
