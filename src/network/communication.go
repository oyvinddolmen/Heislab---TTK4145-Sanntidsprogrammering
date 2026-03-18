package network

import (
	"heislab/management"
	"heislab/state"
	"time"
)

const broadcastInterval = 20 * time.Millisecond

// Initializes communication go routines.
func InitCommunication(
	elev *management.Elevator,
	globalState *state.GlobalState,
	networkChannels NetworkChannels,
) {
	go StartFailureDetector(globalState, networkChannels.WorldViewUpdateChannel)
	go SendGlobalStatePeriodically(elev, globalState, networkChannels.OutgoingGlobalStateChannel)
	go ListenAndMergeGlobalState(
		elev,
		globalState,
		networkChannels.IncomingGlobalStateChannel,
		networkChannels.WorldViewUpdateChannel,
	)
}

// Listens for incoming worldViews, updates globalState and notifies if there is a new world view 
func ListenAndMergeGlobalState(
	elev *management.Elevator,
	globalState *state.GlobalState,
	incomingGlobalStateChannel <-chan state.GlobalStateData,
	worldViewUpdateChannel chan bool,
) {
	localID := globalState.GetLocalID()
	for remoteGlobalState := range incomingGlobalStateChannel {

		// Prevents elev from listening to itself
		if remoteGlobalState.LocalID == globalState.GetCopy().LocalID {
			continue
		}

		RegisterHeartbeat(localID, remoteGlobalState.LocalID)
		if globalState.NewWorldView(remoteGlobalState) {
			if globalState.IsOffline() {
				globalState.SetSelfToOnline(elev)
			}
			globalState.Merge(remoteGlobalState)
			worldViewUpdateChannel <- true
			continue
		}
	}
}

// Sends global state periodically at interval given by broadcastInterval constant
func SendGlobalStatePeriodically(
	elev *management.Elevator,
	globalState *state.GlobalState,
	outgoingGlobalStateChannel chan<- state.GlobalStateData,
) {
	ticker := time.NewTicker(broadcastInterval)
	defer ticker.Stop()

	for range ticker.C {
		SendGlobalState(elev, globalState, outgoingGlobalStateChannel)
	}
}

// Sends global state once
func SendGlobalState(
	elev *management.Elevator,
	globalState *state.GlobalState,
	outgoingGlobalStateChannel chan<- state.GlobalStateData,
) {
	globalState.UpdateGlobalState(elev)
	message := globalState.GetCopy()
	outgoingGlobalStateChannel <- message
}
