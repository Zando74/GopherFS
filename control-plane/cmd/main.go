package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Zando74/GopherFS/control-plane/config"
	"github.com/Zando74/GopherFS/control-plane/internal/application/adapter/consensus"
	"github.com/Zando74/GopherFS/control-plane/internal/application/adapter/consensus/fsm"
	"github.com/Zando74/GopherFS/control-plane/internal/application/adapter/controller"
	"github.com/Zando74/GopherFS/control-plane/logger"
	"github.com/hashicorp/raft"
)

func main() {
	cfg := config.ConfigSingleton.GetInstance()
	log := logger.LoggerSingleton.GetInstance()
	badgerDB := fsm.BabdgerSingleton.GetInstance()
	raftsrv := consensus.RaftServerSingleton.GetInstance()
	controller := controller.Controller{}

	log.Info(logger.ConfigLoadedMessage, cfg)

	leaderElectionChannel := make(chan raft.Observation)
	once := sync.Once{}

	go func() {
		observerLeaderElected := consensus.OnLeaderElection(leaderElectionChannel, cfg.Consensus.NodeId)
		raftsrv.RegisterObserver(observerLeaderElected)

		go func() {
			for range leaderElectionChannel {
				if cfg.Consensus.Leader {
					once.Do(func() {
						consensus.LookingForCandidates(raftsrv)
					})
				}
				controller.OnLeaderElection()
				controller.RecoverPendingSagas()
			}
		}()
	}()

	controller.Run()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	badgerDB.Close()
	raftsrv.Shutdown().Error()
	controller.Shutdown()
	os.Exit(0)

}
