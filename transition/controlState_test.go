////////////////////////////////////////////////////////////////////////////////
// Copyright © 2018 Privategrity Corporation                                   /
//                                                                             /
// All rights reserved.                                                        /
////////////////////////////////////////////////////////////////////////////////
package transition

import (
	"gitlab.com/elixxir/primitives/current"
	"gitlab.com/elixxir/primitives/states"
	"reflect"
	"testing"
)

// Tests the valid transition states
func TestTransitions_IsValidTransition(t *testing.T) {
	testTransition := newTransitions()

	var expectedTransition = make([][]bool, current.NUM_STATES, current.NUM_STATES)

	expectedTransition[current.NOT_STARTED] = []bool{false, false, false, false, false, false, false, false}
	expectedTransition[current.WAITING] = []bool{true, false, false, false, false, true, true, false}
	expectedTransition[current.PRECOMPUTING] = []bool{false, true, false, false, false, false, false, false}
	expectedTransition[current.STANDBY] = []bool{false, true, true, false, false, false, false, false}
	expectedTransition[current.REALTIME] = []bool{false, false, false, true, false, false, false, false}
	expectedTransition[current.COMPLETED] = []bool{false, false, false, false, true, false, false, false}
	expectedTransition[current.ERROR] = []bool{true, true, true, true, true, true, false, false}
	expectedTransition[current.CRASH] = make([]bool, current.NUM_STATES)

	for i := uint32(0); i < uint32(current.NUM_STATES); i++ {
		receivedTransitions := make([]bool, len(expectedTransition))

		for k := uint32(0); k < uint32(current.NUM_STATES); k++ {
			//fmt.Printf("iter %d: %v\n", i, current.Activity(i))
			ok := testTransition.IsValidTransition(current.Activity(i), current.Activity(k))
			receivedTransitions[k] = ok
		}

		if !reflect.DeepEqual(expectedTransition[current.Activity(i)], receivedTransitions) {
			t.Errorf("State transitions for %s did not match expected.\n\tExpected: %v\n\tReceived: %v",
				current.Activity(i), expectedTransition[current.Activity(i)], receivedTransitions)
		}

	}

}

// Checks the look up function for NeedsRound produces expected results
func TestTransitions_NeedsRound(t *testing.T) {
	testTransition := newTransitions()

	expectedNeedsRound := []int{0, 0, 1, 1, 1, 1, 2}
	receivedNeedsRound := make([]int, len(expectedNeedsRound))
	for i := uint32(0); i < uint32(current.CRASH); i++ {
		receivedNeedsRound[i] = testTransition.NeedsRound(current.Activity(i))
	}

	if !reflect.DeepEqual(expectedNeedsRound, receivedNeedsRound) {
		t.Errorf("NeedsRound did not match expected.\n\tExpected: %v\n\tReceived: %v",
			expectedNeedsRound, receivedNeedsRound)
	}
}

// Tests the look up function for RequiredRoundState produces expected results
func TestTransitions_RequiredRoundState(t *testing.T) {
	testTransition := newTransitions()

	input := []states.Round{states.PENDING, states.PENDING, states.PRECOMPUTING,
		states.PRECOMPUTING, states.QUEUED, states.REALTIME, states.PENDING}
	receivedRoundState := make([]bool, len(input))
	expectedRoundState := []bool{false, false, true, true, true, true, false}

	for i := uint32(0); i < uint32(current.CRASH); i++ {
		receivedRoundState[i] = testTransition.IsValidRoundState(current.Activity(i), input[i])
	}

	if !reflect.DeepEqual(expectedRoundState, receivedRoundState) {
		t.Errorf("NeedsRound did not match expected.\n\tExpected: %v\n\tReceived: %v",
			expectedRoundState, receivedRoundState)
	}

}
