package faultTolerance

import (
	"heislab/management"
	"strconv"
	"heislab/orderManagement"
)

// Called once when elevator boots
func RecoverOnStartup() {

	orderManagement.GlobalStateMutex.Lock()
	defer orderManagement.GlobalStateMutex.Unlock()

	localID := strconv.Itoa(management.Elev.ID)

	// Restore cab orders if available
	oldState, exists := orderManagement.GlobalState.States[localID]
	if exists {
		for f := 0; f < management.NumFloors; f++ {
			if oldState.CabRequests[f] {
				management.Elev.Orders[f][2].OrderPlaced = true
			}
		}
	}

	// Clear local hall orders (they are global responsibility)
	clearLocalHallOrders()

	// Mark self alive in globalState (behavior idle)
	orderManagement.GlobalState.States[localID] = orderManagement.ConvertElevatorToJSON(management.Elev)

	// Trigger hall reassignment
	go orderManagement.RunHallAssigner()
}

// Clear local hall orders (only hall buttons, keep cab orders)
func clearLocalHallOrders() {

	for f := 0; f < management.NumFloors; f++ {
		for btn := 0; btn < 2; btn++ { // hall buttons only
			management.Elev.Orders[f][btn].OrderPlaced = false
			management.Elev.Orders[f][btn].ElevID = -1
		}
	}
}
