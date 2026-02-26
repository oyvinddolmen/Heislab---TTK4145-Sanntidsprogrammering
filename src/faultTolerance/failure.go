package faultTolerance

import (
	"heislab/management"
	"heislab/orderManagement"
	"sync"
	"time"
	//"os"
	//"fmt"
)


const HeartbeatTimeout = 2 * time.Second

// Track last time we heard from each elevator
var lastSeen = make(map[string]time.Time)
var failureMutex sync.Mutex

// Called whenever we receive state from another elevator
func RegisterHeartbeat(id string) {
	localID := management.Elev.ID
	if id == localID {
		// Ignore self
		return
	}
	lastSeen[id] = time.Now()
}

// Periodically check if elevators have died
func StartFailureDetector() {

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		checkForDeadElevators()
	}
}

// Detect and handle dead elevators
func checkForDeadElevators() {
	failureMutex.Lock()
	defer failureMutex.Unlock()

	now := time.Now()
	localID := management.Elev.ID

	for id, t := range lastSeen {

		if id == localID { // we do not delete ourself
			continue
		}

		if now.Sub(t) > HeartbeatTimeout {

			handleElevatorFailure(id)
			delete(lastSeen, id)
			handleElevatorFailure(id)
			delete(lastSeen, id)
		}
	}
}

// Set dead elevator behavior to "offline" and redistribute orders
func handleElevatorFailure(deadID string) {

	orderManagement.GlobalStateMutex.Lock()

	state, exists := orderManagement.GlobalState.States[deadID]
	if !exists {
		orderManagement.GlobalStateMutex.Unlock()
		return
	}

	state.Behavior = "offline"
	orderManagement.GlobalState.States[deadID] = state

	orderManagement.GlobalStateMutex.Unlock()

	releaseHallOrders(deadID)

	orderManagement.RunHallAssigner()

	// broadcast her via channel
}

// Release hall orders belonging to dead elevator
func releaseHallOrders(deadID string) {

	orderManagement.GlobalStateMutex.Lock()
	defer orderManagement.GlobalStateMutex.Unlock()

	for f := 0; f < management.NumFloors; f++ {
		for btn := 0; btn < 2; btn++ {

			if orderManagement.GlobalState.HallRequestsAssigned[f][btn] == deadID {

				// sett tilbake som aktiv request
				orderManagement.GlobalState.HallRequests[f][btn] = true

				// fjern assignment
				orderManagement.GlobalState.HallRequestsAssigned[f][btn] = ""
			}
		}
	}
}

/*
func GetStartupInput() (int, error) {

	if len(os.Args) != 2 {
		return 0, fmt.Errorf("usage: go run main.go <ID>")
	}

	elevID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		return 0, fmt.Errorf("ID must be an integer")
	}

	return elevID, nil
}
*/