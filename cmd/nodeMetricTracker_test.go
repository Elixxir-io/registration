////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package cmd

import (
	"bytes"
	"crypto/rand"
	"gitlab.com/elixxir/registration/scheduling"
	"gitlab.com/elixxir/registration/storage"
	"gitlab.com/elixxir/registration/storage/node"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/ndf"
	"gitlab.com/xx_network/primitives/region"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTrackNodeMetrics(t *testing.T) {
	kill := make(chan struct{})
	defer quit(kill)
	interval := 500 * time.Millisecond

	var err error
	storage.PermissioningDb, _, err = storage.NewDatabase("", "", "perm", "", "")
	if err != nil {
		t.Errorf(err.Error())
	}

	testParams.pruneRetentionLimit = 24 * time.Hour
	testParams.disableNDFPruning = false
	// Create a new state
	state, err := storage.NewState(getTestKey(), 8, "", "", region.GetCountryBins())
	if err != nil {
		t.Errorf("Unable to create state: %+v", err)
	}

	// Construct an active node
	activeNodeID := id.NewIdFromString("active", id.Node, t)
	err = state.GetNodeMap().AddNode(activeNodeID, "", "", "", 0)
	if err != nil {
		t.Fatalf("TestTrackNodeMetrics: Failed to add node to state: %v", err)
	}
	activeNode := state.GetNodeMap().GetNode(activeNodeID)
	activeNode.SetNumPollsTesting(25, t)
	activeNode.SetLastActiveTesting(time.Now().Add(interval*2), t)

	// Construct a stale node in the map
	staleNodeId := id.NewIdFromString("stale", id.Node, t)
	err = state.GetNodeMap().AddNode(staleNodeId, "", "", "", 0)
	if err != nil {
		t.Fatalf("TestTrackNodeMetrics: Failed to add node to state: %v", err)
	}
	staleNode := state.GetNodeMap().GetNode(staleNodeId)
	staleNode.GetAndResetNumPolls() // Set to zero
	staleNode.SetLastActiveTesting(time.Now().Add(-interval*3), t)

	// Construct a node which will be pruned
	pruneNodeId := id.NewIdFromString("prune", id.Node, t)
	err = state.GetNodeMap().AddNode(pruneNodeId, "", "", "", 0)
	if err != nil {
		t.Fatalf("TestTrackNodeMetrics: Failed to add node to state: %v", err)
	}
	pruneNode := state.GetNodeMap().GetNode(staleNodeId)
	pruneTimestamp, err := time.Parse(time.RFC3339,
		"2012-12-21T22:08:41+00:00")
	if err != nil {
		t.Fatalf("Could not parse precanned time: %v", err.Error())
	}
	pruneNode.SetLastActiveTesting(pruneTimestamp, t)
	// Add all nodes to database
	nodeIds := []*id.ID{activeNodeID, staleNodeId, pruneNodeId}
	status := []node.Status{node.Active, node.Inactive, node.Inactive}
	for i := 1; i < 4; i++ {
		regCode := strconv.Itoa(i)
		//nid := createNode(state, strconv.Itoa(i), regCode, i, status[i], t)

		// Create random bytes so application Ids don't collide
		idBytes := make([]byte, id.ArrIDLen)
		_, err := rand.Read(idBytes)
		if err != nil {
			t.Fatalf("Failed to generate random bytes: %v", err)
		}

		// Set up reg code
		appId := uint64(i * 10)
		err = storage.PermissioningDb.InsertApplication(
			&storage.Application{Id: appId}, &storage.Node{
				Code:          regCode,
				Id:            idBytes,
				ApplicationId: appId,
				Status:        uint8(status[i-1]),
				Sequence:      strconv.Itoa(i),
			})
		if err != nil {
			t.Fatalf("Failed to insert application: %+v", err)
		}
		err = storage.PermissioningDb.RegisterNode(nodeIds[i-1], nil, regCode, "", "", "", "")
		if err != nil {
			t.Fatalf("Failed to prepopulate database: %+v", err)
		}
	}

	// Construct an NDF with these nodes
	testNdf := &ndf.NetworkDefinition{
		Nodes: []ndf.Node{
			{
				ID: activeNodeID.Bytes(),
			},
			{
				ID: staleNodeId.Bytes(),
			},
			{
				ID: pruneNodeId.Bytes(),
			},
		},
		Gateways: []ndf.Gateway{
			{
				ID: activeNodeID.Bytes(),
			},
			{
				ID: staleNodeId.Bytes(),
			},
			{
				ID: pruneNodeId.Bytes(),
			},
		},
	}

	state.UpdateInternalNdf(testNdf)

	impl := &RegistrationImpl{
		params:               &testParams,
		State:                state,
		earliestRoundTracker: atomic.Value{},
	}

	impl.schedulingParams = &scheduling.SafeParams{
		RWMutex: sync.RWMutex{},
		Params:  &scheduling.Params{},
	}

	go TrackNodeMetrics(impl, kill,
		interval)

	time.Sleep(interval * 4)

	resultNdf := impl.State.GetFullNdf().Get()

	if len(resultNdf.Nodes) != 1 {
		t.Fatalf("Unexpected amount of nodes in NDF."+
			"\n\tExpected: %d"+
			"\n\tReceived: %d", 1, len(resultNdf.Nodes))
	}

	for _, n := range resultNdf.Nodes {
		if bytes.Equal(pruneNodeId.Bytes(), n.ID) {
			t.Fatalf("Pruned node should not be in the NDF")
		} else if bytes.Equal(staleNodeId.Bytes(), n.ID) {
			if n.Status != ndf.Stale {
				t.Fatalf("Stale node has unexpected value"+
					"\n\tExpected: %s"+
					"\n\tReceived: %s", ndf.Stale, n.Status)
			}
		}
	}

}

func quit(kill chan struct{}) {
	kill <- struct{}{}
}
