package network

import (
	"fmt"
	"heislab/management"
	"heislab/orderManagement"
	"sync"
	"time"
	"heislab/hallRequestAssigner"
	"heislab/elevator/elevio"
)

const HeartbeatTimeout = 2 * time.Second

// Track last time we heard from each elevator
var (
	lastSeen = make(map[string]time.Time)
	mu       sync.Mutex
)

// Called whenever we receive state from another elevator
func RegisterHeartbeat(elevID string) {
	localID := management.Elev.ID
	if elevID == "" || elevID == localID {
		// Ignore empty/self IDs
		return
	}

	mu.Lock()
	lastSeen[elevID] = time.Now()
	mu.Unlock()
}

// Periodically check if elevators have died
func StartFailureDetector(gs *orderManagement.GlobalState, worldViewUpdate chan bool) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		checkForDeadElevators(gs, worldViewUpdate)
	}
}

// Detect and handle dead elevators
func checkForDeadElevators(gs *orderManagement.GlobalState, worldViewUpdate chan bool) {
	now := time.Now()
	localID := management.Elev.ID

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
func RecoverOnStartup(gs *orderManagement.GlobalState, rx <-chan orderManagement.GlobalStateType) bool {
	elevID := management.Elev.ID
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
			management.Elev.Floor = recovered.Floor
			management.Elev.LastFloor = recovered.Floor
			elevio.SetFloorIndicator(recovered.Floor)
		}

		for floor := 0; floor < management.NumFloors && floor < len(recovered.CabRequests); floor++ {
			if recovered.CabRequests[floor] {
				management.Elev.Orders[floor][elevio.CabButton].OrderPlaced = true
				management.Elev.Orders[floor][elevio.CabButton].ElevID = elevID
				elevio.SetButtonLamp(elevio.CabButton, floor, true)
			}
		}
	}

	// Register local elevator state in global state after recovery.
	gs.SetElevatorGlobalState(elevID, orderManagement.ConvertElevatorToJSON(management.Elev))
	return true
}

// Listens for incomming worldViews, updates globalState and sends on worldView-channel
func ListenAndMergeGlobalState(gs *orderManagement.GlobalState, rx <-chan orderManagement.GlobalStateType, worldViewUpdate chan bool) {
	for remoteGlobalState := range rx {

		// to prevent elev from listening to itself
		if remoteGlobalState.LocalID == management.Elev.ID {
			continue
		}

		RegisterHeartbeat(remoteGlobalState.LocalID)
		if gs.NewWorldViev(remoteGlobalState) {
			gs.Merge(remoteGlobalState) // need to merge global view before sending on worldViewupdate for lights to be correct
			worldViewUpdate <- true
			continue
		}
	}
}

// Periodically sends global state
func SendGlobalStatePeriodically(gs *orderManagement.GlobalState, tx chan<- orderManagement.GlobalStateType, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		gs.UpdateGlobalState() // oppdater egen state
		msg := gs.GetCopy()    // ta sikker kopi under mutex
		tx <- msg              // send
	}
}

// Sends global state once
func SendGlobalState(gs *orderManagement.GlobalState, tx chan<- orderManagement.GlobalStateType) {
	gs.UpdateGlobalState()
	msg := gs.GetCopy()
	tx <- msg
}
