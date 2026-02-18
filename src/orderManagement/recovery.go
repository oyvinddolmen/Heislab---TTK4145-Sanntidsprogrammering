package orderManagement

import (
	"strconv"
	"heislab/management"
)

// Called once when elevator boots
func RecoverOnStartup() {

	mutex.Lock()
	defer mutex.Unlock()

	localID := strconv.Itoa(management.Elev.ID)

	// Restore cab orders if available
	oldState, exists := globalState.States[localID]
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
	globalState.States[localID] = convertElevatorToJSON(management.Elev)

	// Trigger hall reassignment
	go RunHallAssigner()
}

// Clear local hall orders (only hall buttons, keep cab orders)
func clearLocalHallOrders() {

	for f := 0; f < management.NumFloors; f++ {
		for btn := 0; btn < 2; btn++ { // hall buttons only
			management.Elev.Orders[f][btn].OrderPlaced = false
			management.Elev.Orders[f][btn].Status = -1
		}
	}
}
