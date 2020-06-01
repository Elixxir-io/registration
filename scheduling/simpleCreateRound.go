////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////
package scheduling

import (
	"github.com/pkg/errors"
	"gitlab.com/elixxir/comms/connect"
	"gitlab.com/elixxir/crypto/shuffle"
	"gitlab.com/elixxir/primitives/id"
	"gitlab.com/elixxir/registration/storage"
	"gitlab.com/elixxir/registration/storage/node"
	"strconv"
	"time"
)

// createSimpleRound.go contains the logic to construct a team for a round and
// add that round to the network state

// createSimpleRound.go builds a team for a round out of a pool and round id and places
// this round into the network state
func createSimpleRound(params Params, pool *waitingPool, roundID id.Round,
	state *storage.NetworkState) (protoRound, error) {
	//remove any offline nodes from consideration
	pool.CleanOfflineNodes(time.Duration(params.NodeCleanUpInterval) * time.Minute)

	nodes, err := pool.PickNRandAtThreshold(int(params.TeamSize), int(params.TeamSize))
	if err != nil {
		return protoRound{}, errors.Errorf("Failed to pick random node group: %v", err)
	}

	var newRound protoRound

	//build the topology
	nodeMap := state.GetNodeMap()
	nodeStateList := make([]*node.State, params.TeamSize)
	orderedNodeList := make([]*id.ID, params.TeamSize)

	// In the case of random ordering
	if params.RandomOrdering {

		// Input an incrementing array of ints
		randomIndex := make([]uint64, params.TeamSize)
		for i := range randomIndex {
			randomIndex[i] = uint64(i)
		}

		// Shuffle array of ints randomly using Fisher-Yates shuffle
		// https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
		shuffle.Shuffle(&randomIndex)
		for i, nid := range nodes {
			n := nodeMap.GetNode(nid.GetID())
			nodeStateList[i] = n
			// Use the shuffled array as an indexing order for
			//  the nodes' topological order
			orderedNodeList[randomIndex[i]] = nid.GetID()
		}
	} else {
		// Otherwise go in the order derived
		// from the pool picking and the node's ordering
		for i, nid := range nodes {
			n := nodeMap.GetNode(nid.GetID())
			nodeStateList[i] = n

			position, err := strconv.Atoi(n.GetOrdering())
			if err != nil {
				return protoRound{}, errors.WithMessagef(err,
					"Could not parse ordering info ('%s') from node %s",
					n.GetOrdering(), nid.GetID().String())
			}

			orderedNodeList[position] = nid.GetID()
		}
	}

	// Construct the protoround object
	newRound.Topology = connect.NewCircuit(orderedNodeList)
	newRound.ID = roundID
	newRound.BatchSize = params.BatchSize
	newRound.NodeStateList = nodeStateList
	return newRound, nil

}