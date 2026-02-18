package orderManagment

import (
	"strconv"
	"sync"
	"time"
	"heislab/managment"
)

const HeartbeatTimeout = 2 * time.Second

// Track last time we heard from each elevator
var lastSeen = make(map[string]time.Time)
var failureMutex sync.Mutex

// Called whenever we receive state from another elevator
func RegisterHeartbeat(id string) {
	localID := strconv.Itoa(managment.Elev.ID)
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
	localID := strconv.Itoa(managment.Elev.ID)

	for id, t := range lastSeen {

		if id == localID {    // we do not delete ourself
			continue
		}

		if now.Sub(t) > HeartbeatTimeout {

			handleElevatorFailure(id)
			delete(lastSeen, id)
		}
	}
}

// Remove dead elevator from global state and redistribute orders
func handleElevatorFailure(deadID string) {

	state, exists := globalState.States[deadID]
	if !exists {
		return
	}

	state.Behavior = "offline"
	globalState.States[deadID] = state

	releaseHallOrders(deadID)

	go RunHallAssigner()
}


// Release hall orders belonging to dead elevator
func releaseHallOrders(deadID string) {

	idInt, _ := strconv.Atoi(deadID)

	for f := 0; f < managment.NumFloors; f++ {
		for btn := 0; btn < 2; btn++ { // hall buttons only

			order := &managment.Elev.Orders[f][btn]

			if order.Status == idInt {
				order.Status = -1  // sets order as not handled
				order.OrderPlaced = true //sets order as placed.. need this?
			}
		}
	}
}
