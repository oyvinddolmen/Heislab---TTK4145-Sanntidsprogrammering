package faultTolerance

import (
	"fmt"
	"heislab/management"
	"heislab/orderManagement"
	"sync"
	"time"
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
func StartFailureDetector(gs *orderManagement.GlobalState) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		checkForDeadElevators(gs)
	}
}

// Detect and handle dead elevators
func checkForDeadElevators(gs *orderManagement.GlobalState) {
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
			// Sett heisen offline i globalState
			gs.SetElevatorToOffline(id)

			// Fjern fra lastSeen
			delete(lastSeen, id)

			// Re-kjor hall assigner med oppdatert state
			orderManagement.RunHallAssignerAndApplyAssignments(gs)
		}
	}
}
