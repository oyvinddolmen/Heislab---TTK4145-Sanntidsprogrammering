package network

import (
	"heislab/management"
	"time"
	"heislab/state"
)

const broadcastInterval = 20 * time.Millisecond

// Listens for incoming worldViews, updates globalState and sends on worldView-channel
func ListenAndMergeGlobalState(globalState *state.GlobalState, globalStateRx <-chan state.GlobalStateData, worldViewUpdate chan bool) {
	localID := globalState.GetLocalID()
	for remoteGlobalState := range globalStateRx {

		// Prevents elev from listening to itself
		if remoteGlobalState.LocalID == globalState.GetCopy().LocalID{
			continue
		}

		RegisterHeartbeat(localID, remoteGlobalState.LocalID)
		if globalState.NewWorldView(remoteGlobalState) {
			globalState.Merge(remoteGlobalState) // Need to merge global view before sending on worldViewupdate for lights to be correct
			worldViewUpdate <- true
			continue
		}
	}
}

func SendGlobalStatePeriodically(elev *management.Elevator, globalState *state.GlobalState, globalStateTx chan<- state.GlobalStateData) {
	ticker := time.NewTicker(broadcastInterval)
	defer ticker.Stop()

	for range ticker.C {
		globalState.UpdateGlobalState(elev) // oppdater egen state
		msg := globalState.GetCopy()        // ta sikker kopi under mutex
		globalStateTx <- msg                // send
	}
}

// Sends global state once
func SendGlobalState(elev *management.Elevator, globalState *state.GlobalState, globalStateTx chan<- state.GlobalStateData) {
	globalState.UpdateGlobalState(elev)
	msg := globalState.GetCopy()
	globalStateTx <- msg
}
