package consensus

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Zando74/GopherFS/control-plane/config"
	"github.com/Zando74/GopherFS/control-plane/internal/application/adapter/consensus/fsm"
	"github.com/Zando74/GopherFS/control-plane/logger"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

const (
	maxPool = 3

	tcpTimeout = 10 * time.Second

	raftSnapShotRetain = 2

	raftLogCacheSize = 512
)

func StartRaftCluster() *raft.Raft {
	configInstance := config.ConfigSingleton.GetInstance()
	log := logger.LoggerSingleton.GetInstance()
	badgerDB := fsm.BabdgerSingleton.GetInstance()

	var raftBinAddr = fmt.Sprintf(":%d", configInstance.Consensus.RaftPort)
	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(configInstance.Consensus.NodeId)
	raftConf.SnapshotThreshold = 1024

	fsmStore := fsm.NewBadger(badgerDB)
	store, err := raftboltdb.NewBoltStore(filepath.Join(configInstance.Consensus.RaftDir, "raft.dataRepo"))
	if err != nil {
		log.Fatal(err)
	}

	cacheStore, err := raft.NewLogCache(raftLogCacheSize, store)
	if err != nil {
		log.Fatal(err)
	}

	snapshotStore, err := raft.NewFileSnapshotStore(configInstance.Consensus.RaftDir, raftSnapShotRetain, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost"+raftBinAddr)
	if err != nil {
		log.Fatal(err)
	}

	transport, err := raft.NewTCPTransport(raftBinAddr, tcpAddr, maxPool, tcpTimeout, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	raftServer, err := raft.NewRaft(raftConf, fsmStore, cacheStore, store, snapshotStore, transport)
	if err != nil {
		log.Fatal(err)
	}

	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(configInstance.Consensus.NodeId),
				Address: transport.LocalAddr(),
			},
		},
	}

	raftServer.BootstrapCluster(configuration)

	return raftServer

}

func LookingForFollowers(raftsrv *raft.Raft) {
	cfg := config.ConfigSingleton.GetInstance()
	go func() {
		time.Sleep(5 * time.Second)
		for _, follower := range cfg.Consensus.Followers {
			followerID := raft.ServerID(follower)
			followerAddress := raft.ServerAddress(fmt.Sprintf("%s:%d", follower, cfg.Consensus.RaftPort))
			if err := raftsrv.AddVoter(followerID, followerAddress, 0, 0).Error(); err != nil {
				logger.LoggerSingleton.GetInstance().Errorf("failed to add voter: %v", err)
			}
		}
	}()
}

func MonitorLeader(raftsrv *raft.Raft) {
	leaderCh := raftsrv.LeaderCh()
	for {
		select {
		case isLeader := <-leaderCh:
			if isLeader {
				logger.LoggerSingleton.GetInstance().Info("Node has become the leader")
			} else {
				logger.LoggerSingleton.GetInstance().Info("Node has lost leadership, triggering re-election")
				raftsrv.LeadershipTransfer()
			}
		case <-time.After(5 * time.Second):
			if raftsrv.Leader() == raft.ServerAddress("") {
				logger.LoggerSingleton.GetInstance().Info("Leader is not reachable, triggering re-election")
				RaftServerSingleton.Reinitialize()
			}
		}
	}
}

type RaftServer struct {
	once sync.Once
	raft *raft.Raft
}

func (rs *RaftServer) GetInstance() *raft.Raft {
	rs.once.Do(func() {
		rs.raft = StartRaftCluster()
	})
	return rs.raft
}

func (rs *RaftServer) Reinitialize() {
	rs.raft.Shutdown()
	rs.raft = StartRaftCluster()
}

var RaftServerSingleton RaftServer
