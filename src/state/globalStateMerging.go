package state

import (
	"heislab/management"
)

// checks if incoming worldView differ from current worldView
func (globalState *GlobalState) NewWorldView(remote GlobalStateData) bool {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()

	// hall request changes
	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < 2; button++ {
			localVersion := globalState.data.HallRequestsVersion[floor][button]
			remoteVersion := remote.HallRequestsVersion[floor][button]

			if remoteVersion > localVersion {
				return true
			}

			// same version but remote has request we don't
			if remoteVersion == localVersion &&
				remote.HallRequests[floor][button] &&
				!globalState.data.HallRequests[floor][button] {
				return true
			}
		}
	}

	// elevator state changes
	senderID := remote.LocalID
	if senderID == "" || senderID == globalState.data.LocalID {
		return false
	}

	remoteState, ok := remote.States[senderID]
	if !ok {
		return false
	}

	localState, exists := globalState.data.States[senderID]
	if !exists {
		return true
	}

	if remoteState.CanTakeOrders != localState.CanTakeOrders {
		return true
	}

	if remoteState.Behavior != localState.Behavior ||
		remoteState.Floor != localState.Floor ||
		remoteState.Direction != localState.Direction {
		return true
	}

	if len(remoteState.CabRequests) != len(localState.CabRequests) {
		return true
	}

	for i := range remoteState.CabRequests {
		if remoteState.CabRequests[i] != localState.CabRequests[i] {
			return true
		}
	}

	return false
}

// merges new global view by choosing newest version available
func (globalState *GlobalState) Merge(remote GlobalStateData) {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()

	localID := globalState.data.LocalID
	senderID := remote.LocalID

	if senderID != localID {
		if st, exists := remote.States[senderID]; exists {
			globalState.data.States[senderID] = st
		}
	}
	chooseLatestHallRequestVersion(&globalState.data, remote)
}

func chooseLatestHallRequestVersion(local *GlobalStateData, remote GlobalStateData) {
	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < 2; button++ {
			localVersion := local.HallRequestsVersion[floor][button]
			remoteVersion := remote.HallRequestsVersion[floor][button]

			switch {
			case remoteVersion > localVersion:
				local.HallRequests[floor][button] = remote.HallRequests[floor][button]
				local.HallRequestsVersion[floor][button] = remoteVersion
			case remoteVersion == localVersion:
				if remote.HallRequests[floor][button] {
					local.HallRequests[floor][button] = true
				}
			}
		}
	}
}
