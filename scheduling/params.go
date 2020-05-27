////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////
package scheduling

// Contains the scheduling params object and the internal protoround object

import (
	"gitlab.com/elixxir/comms/connect"
	"gitlab.com/elixxir/primitives/id"
	"gitlab.com/elixxir/registration/storage/node"
	"time"
)

// JSONable structure which defines the parameters of the scheduler
type Params struct {
	// selects if the secure or simple node selection algorithm is used
	Secure bool
	// number of nodes in a team
	TeamSize uint32
	// number of slots in a batch
	BatchSize uint32

	// Time in ms between assigning a round
	MinimumDelay time.Duration
	// delay in ms for a realtime round to start
	RealtimeDelay time.Duration
	// Time in minutes between cleaning up offline nodes
	NodeCleanUpInterval time.Duration

	//SIMPLE ONLY//
	// sets if simple teaming randomly orders nodes or orders based upon the
	// number in the `order` string
	RandomOrdering bool

	//SECURE ONLY
	// sets the minimum number of nodes in the waiting pool before secure teaming
	// wil create a team
	Threshold uint32
}

//internal structure which describes a round to be created
type protoRound struct {
	Topology      *connect.Circuit
	ID            id.Round
	NodeStateList []*node.State
	BatchSize     uint32
}
