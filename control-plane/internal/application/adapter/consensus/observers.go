package consensus

import (
	"github.com/hashicorp/raft"
)

func OnLeaderElection(leaderElectionChannel chan raft.Observation, nodeId string) *raft.Observer {
	return raft.NewObserver(
		leaderElectionChannel,
		false,
		func(o *raft.Observation) bool {
			if _, ok := o.Data.(raft.LeaderObservation); ok {
				if o.Data.(raft.LeaderObservation).LeaderID == raft.ServerID(nodeId) {
					leaderElectionChannel <- *o
				}

				return true
			}
			return false
		},
	)
}
