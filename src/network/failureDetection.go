package network

import (
	"fmt"
	"heislab/state"
	"sync"
	"time"
)

const HeartbeatTimeout = 2 * time.Second

// Track last time we heard from each elevator.
var (
	lastSeen = make(map[string]time.Time)
	mutex sync.Mutex
)

// Called whenever we receive state from another elevator.
func RegisterHeartbeat(localID string, remoteElevID string) {
	if remoteElevID == "" || remoteElevID == localID { // Ignore empty/self IDs.
		return
	}
	mutex.Lock()
	lastSeen[remoteElevID] = time.Now()
	mutex.Unlock()
}

// Periodically checks if elevators have died.
func StartFailureDetector(globalState *state.GlobalState, worldViewUpdateChannel chan bool) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		checkForDeadElevators(globalState, worldViewUpdateChannel)
	}
}

// Detects and handles dead elevators.
func checkForDeadElevators(globalState *state.GlobalState, worldViewUpdateChannel chan bool) {
	now := time.Now()
	localID := globalState.GetLocalID()

	mutex.Lock()
	defer mutex.Unlock()

	for elevID, t := range lastSeen {
		if elevID == localID { // We do not delete ourself
			continue
		}

		if now.Sub(t) > HeartbeatTimeout {
			fmt.Println("elevator", elevID, "went offline")
			globalState.SetElevatorToOffline(elevID)
			if globalState.AllOtherElevatorsOffline(){
				globalState.SetElevatorToOffline(localID)
			}
			delete(lastSeen, elevID)
			worldViewUpdateChannel <- true
		}
	}
}

