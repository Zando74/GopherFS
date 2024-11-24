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
	raftConf.ElectionTimeout = 20 * time.Second
	raftConf.HeartbeatTimeout = 10 * time.Second
	raftConf.LeaderLeaseTimeout = 5 * time.Second
	raftConf.PreVoteDisabled = true

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

func (rs *RaftServer) FetchBadgerDBFromFollower(followerAddress string) error {
	cfg := config.ConfigSingleton.GetInstance()
	raftServer := rs.GetInstance()

	followerID := raft.ServerID(followerAddress)
	follower := raft.ServerAddress(fmt.Sprintf("%s:%d", followerAddress, cfg.Consensus.RaftPort))

	if err := raftServer.AddNonvoter(followerID, follower, 0, 0).Error(); err != nil {
		return err
	}

	return nil
}

func LookingForCandidates(raftsrv *raft.Raft) {
	cfg := config.ConfigSingleton.GetInstance()
	go func() {
		for _, voter := range cfg.Consensus.Candidates {
			voterID := raft.ServerID(voter)
			voterAddress := raft.ServerAddress(fmt.Sprintf("%s:%d", voter, cfg.Consensus.RaftPort))

			if err := raftsrv.AddVoter(voterID, voterAddress, 0, 0).Error(); err != nil {
				logger.LoggerSingleton.GetInstance().Errorf("failed to add voter: %v", err)
			}
		}
	}()
}

type RaftServer struct {
	once sync.Once
	raft *raft.Raft
}

func (rs *RaftServer) ReinstateRaftCluster() {
	rs.raft = StartRaftCluster()
}

func (rs *RaftServer) GetInstance() *raft.Raft {
	rs.once.Do(func() {
		rs.raft = StartRaftCluster()
	})
	return rs.raft
}

var RaftServerSingleton RaftServer
