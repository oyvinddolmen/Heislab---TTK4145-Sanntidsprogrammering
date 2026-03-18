package state

import (
	"heislab/management"
)

// Checks if incoming worldView differs from current worldView.
func (localGlobalState *GlobalState) NewWorldView(remoteGlobalState GlobalStateData) bool {
	localGlobalState.mutex.Lock()
	defer localGlobalState.mutex.Unlock()

	// Checks hall order differences.
	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < management.NumHallButtonTypes; button++ {
			localVersion := localGlobalState.data.HallOrderVersion[floor][button]
			remoteVersion := remoteGlobalState.HallOrderVersion[floor][button]

			if remoteVersion > localVersion {
				return true
			}

			// Same version, but remote has orders we don't.
			if remoteVersion == localVersion &&
				remoteGlobalState.HallOrders[floor][button] &&
				!localGlobalState.data.HallOrders[floor][button] {
				return true
			}
		}
	}

	remoteID := remoteGlobalState.LocalID
	if remoteID == "" || remoteID == localGlobalState.data.LocalID {
		return false
	}

	remoteState, existsInRemote := remoteGlobalState.States[remoteID]
	if !existsInRemote {
		return false
	}

	localRemoteState, existsInLocal := localGlobalState.data.States[remoteID]
	if !existsInLocal {
		return true
	}

	if remoteState.CanTakeOrders != localRemoteState.CanTakeOrders {
		return true
	}

	if remoteState.Behavior != localRemoteState.Behavior ||
		remoteState.Floor != localRemoteState.Floor ||
		remoteState.Direction != localRemoteState.Direction {
		return true
	}

	if len(remoteState.CabOrders) != len(localRemoteState.CabOrders) {
		return true
	}

	for floor := range remoteState.CabOrders {
		if remoteState.CabOrders[floor] != localRemoteState.CabOrders[floor] {
			return true
		}
	}

	return false
}

// Merges new global state by choosing newest version available.
func (localGlobalState *GlobalState) Merge(remoteGlobalState GlobalStateData) {
	localGlobalState.mutex.Lock()
	defer localGlobalState.mutex.Unlock()

	localID := localGlobalState.data.LocalID
	remoteID := remoteGlobalState.LocalID

	if remoteID != localID {
		if remoteState, exists := remoteGlobalState.States[remoteID]; exists {
			localGlobalState.data.States[remoteID] = remoteState
		}
	}
	chooseLatestHallOrderVersion(&localGlobalState.data, remoteGlobalState)
}

func chooseLatestHallOrderVersion(localGlobalState *GlobalStateData, remoteGlobalState GlobalStateData) {
	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < management.NumHallButtonTypes; button++ {
			localVersion := localGlobalState.HallOrderVersion[floor][button]
			remoteVersion := remoteGlobalState.HallOrderVersion[floor][button]

			switch {
			case remoteVersion > localVersion:
				localGlobalState.HallOrders[floor][button] = remoteGlobalState.HallOrders[floor][button]
				localGlobalState.HallOrderVersion[floor][button] = remoteVersion
			case remoteVersion == localVersion:
				if remoteGlobalState.HallOrders[floor][button] {
					localGlobalState.HallOrders[floor][button] = true
				}
			}
		}
	}
}
