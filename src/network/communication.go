package network

import (
	"heislab/management"
	"time"
	"heislab/state"
)

const broadcastInterval = 20 * time.Millisecond

// Listens for incoming worldViews, updates globalState and sends on worldView-channel.
func ListenAndMergeGlobalState(globalState *state.GlobalState, incomingGlobalStateChannel <-chan state.GlobalStateData, worldViewUpdateChannel chan bool) {
	localID := globalState.GetLocalID()
	for remoteGlobalState := range incomingGlobalStateChannel {

		// Prevents elev from listening to itself
		if remoteGlobalState.LocalID == globalState.GetCopy().LocalID{
			continue
		}

		RegisterHeartbeat(localID, remoteGlobalState.LocalID)
		if globalState.NewWorldView(remoteGlobalState) {
			globalState.Merge(remoteGlobalState) // Need to merge global view before sending on worldViewupdate for lights to be correct.
			worldViewUpdateChannel <- true
			continue
		}
	}
}

func SendGlobalStatePeriodically(elev *management.Elevator, globalState *state.GlobalState, outgoingGlobalStateChannel chan<- state.GlobalStateData) {
	ticker := time.NewTicker(broadcastInterval)
	defer ticker.Stop()

	for range ticker.C {					// TODO fjern eller oppdater kommentarer
		globalState.UpdateGlobalState(elev) // oppdater egen state
		msg := globalState.GetCopy()        // ta sikker kopi under mutex
		outgoingGlobalStateChannel <- msg   // send
	}
}

// Sends global state once.
func SendGlobalState(elev *management.Elevator, globalState *state.GlobalState, outgoingGlobalStateChannel chan<- state.GlobalStateData) {
	globalState.UpdateGlobalState(elev)
	msg := globalState.GetCopy()
	outgoingGlobalStateChannel <- msg
}
