package faultTolerance

import (
	"heislab/management"
	"heislab/orderManagement"
	"time"
)

const HeartbeatTimeout = 2 * time.Second

// Track last time we heard from each elevator
var lastSeen = make(map[string]time.Time)

// Called whenever we receive state from another elevator
func RegisterHeartbeat(elevID string) {
	localID := management.Elev.ID
	if elevID == localID {
		// Ignore self
		return
	}
	lastSeen[elevID] = time.Now()
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

	for id, t := range lastSeen {

		if id == localID { // we do not delete ourself
			continue
		}

		if now.Sub(t) > HeartbeatTimeout {
			// Sett heisen offline i globalState
			gs.SetElevatorToOffline(id)

			// Fjern fra lastSeen
			delete(lastSeen, id)

			// Re-kjør hall assigner med oppdatert state
			orderManagement.RunHallAssignerAndApplyAssignments(gs)
		}
	}
}