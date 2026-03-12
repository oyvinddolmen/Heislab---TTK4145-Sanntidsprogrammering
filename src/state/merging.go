package state

import (
	"heislab/management"
)

func (gs *GlobalState) NewWorldView(remote GlobalStateData) bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// hall request changes
	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < 2; button++ {
			localV := gs.data.HallRequestsVersion[floor][button]
			remoteV := remote.HallRequestsVersion[floor][button]

			if remoteV > localV {
				return true
			}

			// same version but remote has request we don't
			if remoteV == localV &&
				remote.HallRequests[floor][button] &&
				!gs.data.HallRequests[floor][button] {
				return true
			}
		}
	}

	// elevator state changes
	senderID := remote.LocalID
	if senderID == "" || senderID == gs.data.LocalID {
		return false
	}

	remoteState, ok := remote.States[senderID]
	if !ok {
		return false
	}

	localState, exists := gs.data.States[senderID]
	if !exists {
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

func (gs *GlobalState) Merge(remote GlobalStateData) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	localID := gs.data.LocalID
	senderID := remote.LocalID

	if senderID != localID {
		if st, exists := remote.States[senderID]; exists {
			gs.data.States[senderID] = st
		}
	}
	chooseLatestHallRequestVersions(&gs.data, remote)
}

func chooseLatestHallRequestVersions(local *GlobalStateData, remote GlobalStateData) {
	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < 2; button++ {
			localV := local.HallRequestsVersion[floor][button]
			remoteV := remote.HallRequestsVersion[floor][button]

			switch {
			case remoteV > localV:
				local.HallRequests[floor][button] = remote.HallRequests[floor][button]
				local.HallRequestsVersion[floor][button] = remoteV
			case remoteV == localV:
				if remote.HallRequests[floor][button] {
					local.HallRequests[floor][button] = true
				}
			}
		}
	}
}
