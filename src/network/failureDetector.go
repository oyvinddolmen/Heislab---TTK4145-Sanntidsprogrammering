package network

import (
	"fmt"
	"heislab/management"
	"sync"
	"time"
	"heislab/hallRequestAssigner"
	"heislab/elevator/elevio"
	"heislab/state"
)

const HeartbeatTimeout = 2 * time.Second

// Track last time we heard from each elevator
var (
	lastSeen = make(map[string]time.Time)
	mu       sync.Mutex
)

// Called whenever we receive state from another elevator
func RegisterHeartbeat(localID string, remoteElevID string) {
	if remoteElevID == "" || remoteElevID == localID {
		// Ignore empty/self IDs
		return
	}
	mu.Lock()
	lastSeen[remoteElevID] = time.Now()
	mu.Unlock()
}

// Periodically check if elevators have died
func StartFailureDetector(gs *state.GlobalState, worldViewUpdate chan bool) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		checkForDeadElevators(gs, worldViewUpdate)
	}
}

// Detect and handle dead elevators
func checkForDeadElevators(gs *state.GlobalState, worldViewUpdate chan bool) {
	now := time.Now()
	localID := gs.GetID()

	mu.Lock()
	defer mu.Unlock()

	for id, t := range lastSeen {
		if id == localID { // we do not delete ourself
			continue
		}

		if now.Sub(t) > HeartbeatTimeout {
			fmt.Println("---------- ELEV", id, "went offline -------------")
			gs.SetElevatorToOffline(id)
			delete(lastSeen, id)
			worldViewUpdate <- true
		}
	}
}

// listens and merge global view received on startup. Returns true if received view 
func RecoverOnStartup(elev *management.Elevator, gs *state.GlobalState, rx <-chan state.GlobalStateData) bool {
	elevID := gs.GetID()
	timeout := time.After(1 * time.Second)
	var recovered *hallRequestAssigner.ElevatorStateJSON

	for {
		select {
		case globalState := <-rx:
			gs.Merge(globalState)

			if st, exists := globalState.States[elevID]; exists {
				tmp := st
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
			elevio.SetFloorIndicator(recovered.Floor)
		}

		for floor := 0; floor < management.NumFloors && floor < len(recovered.CabRequests); floor++ {
			if recovered.CabRequests[floor] {
				elev.Orders[floor][elevio.CabButton].OrderPlaced = true
				elevio.SetButtonLamp(elevio.CabButton, floor, true)
			}
		}
	}

	// Register local elevator state in global state after recovery.
	gs.SetElevatorGlobalState(elevID, state.ConvertElevatorToJSON(elev))
	return true
}