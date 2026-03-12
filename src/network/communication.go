package network

import (
	"heislab/management"
	"time"
	"heislab/state"
)

// Listens for incomming worldViews, updates globalState and sends on worldView-channel
func ListenAndMergeGlobalState(gs *state.GlobalState, rx <-chan state.GlobalStateData, worldViewUpdate chan bool) {
	localID := gs.GetID()
	for remoteGlobalState := range rx {

		// to prevent elev from listening to itself
		if remoteGlobalState.LocalID == gs.GetCopy().LocalID{
			continue
		}

		RegisterHeartbeat(localID, remoteGlobalState.LocalID)
		if gs.NewWorldView(remoteGlobalState) {
			gs.Merge(remoteGlobalState) // need to merge global view before sending on worldViewupdate for lights to be correct
			worldViewUpdate <- true
			continue
		}
	}
}

func SendGlobalStatePeriodically(elev *management.Elevator, gs *state.GlobalState, tx chan<- state.GlobalStateData, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		gs.UpdateGlobalState(elev) // oppdater egen state
		msg := gs.GetCopy()    // ta sikker kopi under mutex
		tx <- msg              // send
	}
}

// Sends global state once
func SendGlobalState(elev *management.Elevator, gs *state.GlobalState, tx chan<- state.GlobalStateData) {
	gs.UpdateGlobalState(elev)
	msg := gs.GetCopy()
	tx <- msg
}
