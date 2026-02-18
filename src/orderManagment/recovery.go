package orderManagment

import (
	"strconv"
	"heislab/managment"
)

// Called once when elevator boots
func RecoverOnStartup() {

	mutex.Lock()
	defer mutex.Unlock()

	localID := strconv.Itoa(managment.Elev.ID)

	// Restore cab orders if available
	oldState, exists := globalState.States[localID]
	if exists {
		for f := 0; f < managment.NumFloors; f++ {
			if oldState.CabRequests[f] {
				managment.Elev.Orders[f][2].OrderPlaced = true
			}
		}
	}

	// Clear local hall orders (they are global responsibility)
	clearLocalHallOrders()

	// Mark self alive in globalState (behavior idle)
	globalState.States[localID] = convertElevatorToJSON(managment.Elev)

	// Trigger hall reassignment
	go RunHallAssigner()
}

// Clear local hall orders (only hall buttons, keep cab orders)
func clearLocalHallOrders() {

	for f := 0; f < managment.NumFloors; f++ {
		for btn := 0; btn < 2; btn++ { // hall buttons only
			managment.Elev.Orders[f][btn].OrderPlaced = false
			managment.Elev.Orders[f][btn].Status = -1
		}
	}
}
