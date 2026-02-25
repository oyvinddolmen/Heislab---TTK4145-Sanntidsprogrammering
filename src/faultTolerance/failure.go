package faultTolerance

import (
	"heislab/management"
	"heislab/orderManagement"
	"sync"
	"time"
)


const HeartbeatTimeout = 2 * time.Second

// Track last time we heard from each elevator
var lastSeen = make(map[string]time.Time)
var failureMutex sync.Mutex

// Called whenever we receive state from another elevator
func RegisterHeartbeat(id string) {
	localIP := management.Elev.IP
	if id == localIP {
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
	localIP := management.Elev.IP

	for ip, t := range lastSeen {

		if ip == localIP { // we do not delete ourself
			continue
		}

		if now.Sub(t) > HeartbeatTimeout {

			handleElevatorFailure(ip)
			delete(lastSeen, ip)
		}
	}
}

// Remove dead elevator from global state and redistribute orders
func handleElevatorFailure(deadIP string) {

	state, exists := orderManagement.GlobalState.States[deadIP]
	if !exists {
		return
	}

	state.Behavior = "offline"
	orderManagement.GlobalState.States[deadIP] = state

	releaseHallOrders(deadIP)

	go orderManagement.RunHallAssigner()
}

// Release hall orders belonging to dead elevator
func releaseHallOrders(deadIP string) {

	for f := 0; f < management.NumFloors; f++ {
		for btn := 0; btn < 2; btn++ { // hall buttons only

			order := &management.Elev.Orders[f][btn]

			if order.ElevIP == deadIP {
				order.ElevIP = ""        // sets order as not handled
				order.OrderPlaced = true //sets order as placed.. need this?
			}
		}
	}
}
