////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////

// Handles network state tracking and control

package storage

import (
	jww "github.com/spf13/jwalterweatherman"
	pb "gitlab.com/elixxir/comms/mixmessages"
	"gitlab.com/elixxir/comms/network/dataStructures"
	"gitlab.com/elixxir/crypto/signature"
	"gitlab.com/elixxir/crypto/signature/rsa"
	"gitlab.com/elixxir/primitives/current"
	"gitlab.com/elixxir/primitives/id"
	"gitlab.com/elixxir/primitives/ndf"
	"gitlab.com/elixxir/primitives/states"
	"sync"
	"sync/atomic"
	"time"
)

// Used for keeping track of NDF and Round state
type State struct {
	// State parameters ---
	PrivateKey *rsa.PrivateKey

	// Round state ---
	CurrentRound  *RoundState
	CurrentUpdate int // Round update counter
	RoundUpdates  *dataStructures.Updates
	RoundData     *dataStructures.Data
	Update        chan struct{} // For triggering updates to top level

	// Node State ---
	NodeStates map[id.Node]*NodeState

	// NDF state ---
	partialNdf *dataStructures.Ndf
	fullNdf    *dataStructures.Ndf
}

// Tracks the current global state of a round
type RoundState struct {
	// Tracks round information
	*pb.RoundInfo

	// Keeps track of the real state of the network
	// as described by the cumulative states of nodes
	// In other words, counts the number of nodes currently in each state
	NetworkStatus [states.NUM_STATES]*uint32
}

// Tracks state of an individual Node in the network
type NodeState struct {
	mux sync.RWMutex

	// Current activity as reported by the Node
	Activity current.Activity

	// Timestamp of the last time this Node polled
	LastPoll time.Time
}

// Returns a new State object
func NewState() (*State, error) {
	fullNdf, err := dataStructures.NewNdf(&ndf.NetworkDefinition{})
	if err != nil {
		return nil, err
	}
	partialNdf, err := dataStructures.NewNdf(&ndf.NetworkDefinition{})
	if err != nil {
		return nil, err
	}

	state := &State{
		CurrentRound: &RoundState{
			RoundInfo: &pb.RoundInfo{
				Topology: make([]string, 0),        // Set this to avoid segfault
				State:    uint32(states.COMPLETED), // Set this to start rounds
			},
		},
		CurrentUpdate: 0,
		RoundUpdates:  dataStructures.NewUpdates(),
		RoundData:     dataStructures.NewData(),
		Update:        make(chan struct{}),
		NodeStates:    make(map[id.Node]*NodeState),
		fullNdf:       fullNdf,
		partialNdf:    partialNdf,
	}

	// Insert dummy update
	err = state.AddRoundUpdate(&pb.RoundInfo{})
	if err != nil {
		return nil, err
	}
	return state, nil
}

// Returns the full NDF
func (s *State) GetFullNdf() *dataStructures.Ndf {
	return s.fullNdf
}

// Returns the partial NDF
func (s *State) GetPartialNdf() *dataStructures.Ndf {
	return s.partialNdf
}

// Returns all updates after the given ID
func (s *State) GetUpdates(id int) ([]*pb.RoundInfo, error) {
	return s.RoundUpdates.GetUpdates(id)
}

// Returns the NodeState object for the given id if it exists
// Otherwise, it will safely create and return a new NodeState
func (s *State) GetNodeState(id id.Node) *NodeState {
	// Create the node state entry if it doesn't already exist
	if s.NodeStates[id] == nil {
		s.NodeStates[id] = &NodeState{}
	}
	return s.NodeStates[id]
}

// Returns true if given node ID is participating in the current round
func (s *State) IsRoundNode(id string) bool {
	for _, nodeId := range s.CurrentRound.Topology {
		if nodeId == id {
			return true
		}
	}
	return false
}

// Returns the state of the current round
func (s *State) GetCurrentRoundState() states.Round {
	return states.Round(s.CurrentRound.State)
}

// Makes a copy of the round before inserting into RoundUpdates
func (s *State) AddRoundUpdate(round *pb.RoundInfo) error {
	roundCopy := &pb.RoundInfo{
		ID:         round.GetID(),
		UpdateID:   round.GetUpdateID(),
		State:      round.GetState(),
		BatchSize:  round.GetBatchSize(),
		Topology:   round.GetTopology(),
		Timestamps: round.GetTimestamps(),
		Signature: &pb.RSASignature{
			Nonce:     round.GetNonce(),
			Signature: round.GetSig(),
		},
	}
	jww.DEBUG.Printf("Round state updated to %s",
		states.Round(roundCopy.State))

	return s.RoundUpdates.AddRound(roundCopy)
}

// Given a full NDF, updates internal NDF structures
func (s *State) UpdateNdf(newNdf *ndf.NetworkDefinition) (err error) {
	// Build NDF comms messages
	fullNdfMsg := &pb.NDF{}
	fullNdfMsg.Ndf, err = newNdf.Marshal()
	if err != nil {
		return
	}
	partialNdfMsg := &pb.NDF{}
	partialNdfMsg.Ndf, err = newNdf.StripNdf().Marshal()
	if err != nil {
		return
	}

	// Sign NDF comms messages
	err = signature.Sign(fullNdfMsg, s.PrivateKey)
	if err != nil {
		return
	}
	err = signature.Sign(partialNdfMsg, s.PrivateKey)
	if err != nil {
		return
	}

	// Assign NDF comms messages
	err = s.fullNdf.Update(fullNdfMsg)
	if err != nil {
		return err
	}
	return s.partialNdf.Update(partialNdfMsg)
}

// Updates the state of the given node with the new state provided
func (s *State) UpdateNodeState(id id.Node, newActivity current.Activity) error {

	// Get and lock node state
	node := s.GetNodeState(id)
	node.mux.Lock()
	defer node.mux.Unlock()

	// Update node poll timestamp
	node.LastPoll = time.Now()

	// Check whether node state requires an update
	if node.Activity != newActivity {
		// Node state was updated, convert node activity to round state
		roundState, err := newActivity.ConvertToRoundState()
		if err != nil {
			return err
		}

		// Update node state tracker
		node.Activity = newActivity

		// Increment network state counter
		atomic.AddUint32(s.CurrentRound.NetworkStatus[roundState], 1)

		// Cue an update
		s.Update <- struct{}{}
	}

	return nil
}
