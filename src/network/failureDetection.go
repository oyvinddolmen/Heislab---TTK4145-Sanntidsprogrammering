package network

import (
	"fmt"
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/state"
	"sync"
	"time"
)

const HeartbeatTimeout = 2 * time.Second

// Track last time we heard from each elevator
var (
	lastSeen = make(map[string]time.Time)
	mutex      sync.Mutex
)

// Called whenever we receive state from another elevator
func RegisterHeartbeat(localID string, remoteElevID string) {
	if remoteElevID == "" || remoteElevID == localID {
		// Ignore empty/self IDs
		return
	}
	mutex.Lock()
	lastSeen[remoteElevID] = time.Now()
	mutex.Unlock()
}

// Periodically check if elevators have died
func StartFailureDetector(globalState *state.GlobalState, worldViewUpdate chan bool) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		checkForDeadElevators(globalState, worldViewUpdate)
	}
}

// Detect and handle dead elevators
func checkForDeadElevators(globalState *state.GlobalState, worldViewUpdate chan bool) {
	now := time.Now()
	localID := globalState.GetLocalID()

	mutex.Lock()
	defer mutex.Unlock()

	for elevID, t := range lastSeen {
		if elevID == localID { // we do not delete ourself
			continue
		}

		if now.Sub(t) > HeartbeatTimeout {
			fmt.Println("---------- ELEV", elevID, "went offline -------------")
			globalState.SetElevatorToOffline(elevID)
			delete(lastSeen, elevID)
			worldViewUpdate <- true
		}
	}
}

// listens and merge global view received on startup. Returns true if received view 
func RecoverOnStartup(elev *management.Elevator, globalState *state.GlobalState, globalStateRx <-chan state.GlobalStateData) bool {
	elevID := globalState.GetLocalID()
	timeout := time.After(1 * time.Second)
	var recovered *state.ElevatorStateJSON

	for {
		select {
		case newGlobalState := <-globalStateRx:
			globalState.Merge(newGlobalState)

			if state, exists := newGlobalState.States[elevID]; exists {
				tmp := state
				recovered = &tmp
			}
		case <-timeout:
			goto RECOVER
		}
	}

RECOVER:
	if recovered == nil {
		fmt.Println("No previous cab state found on startup, starting fresh")
		return false
	} else {
		if recovered.Floor >= 0 && recovered.Floor < management.NumFloors {
			elev.Floor = recovered.Floor
			elev.LastFloor = recovered.Floor
			elevIO.SetFloorIndicator(recovered.Floor)
		}

		for floor := 0; floor < management.NumFloors && floor < len(recovered.CabOrders); floor++ {
			if recovered.CabOrders[floor] {
				elev.Orders[floor][elevIO.CabButton].OrderPlaced = true
				elevIO.SetButtonLamp(elevIO.CabButton, floor, true)
			}
		}
	}

	// Register local elevator state in global state after recovery.
	globalState.SetElevatorGlobalState(elevID, state.ConvertElevatorToJSON(elev))
	return true
}