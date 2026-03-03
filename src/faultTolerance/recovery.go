package faultTolerance

import (
	"fmt"
	"heislab/elevio"
	"heislab/management"
	"heislab/orderManagement"
	"time"
)

// Called once when elevator boots. Checks if other elevs has your cab orders
func RecoverOnStartup(rx <-chan orderManagement.GlobalStateType) {
	timeout := time.After(1 * time.Second)

	// Vent på eksisterende GlobalState
	for {
		select {

		case gs := <-rx:
			orderManagement.GlobalStateMutex.Lock()
			orderManagement.GlobalState = gs
			orderManagement.GlobalStateMutex.Unlock()
			goto RECOVER

		case <-timeout:
			fmt.Println("No GlobalState received on startup, starting fresh")
			goto RECOVER
		}
	}

RECOVER:

	elevID := management.Elev.ID

	orderManagement.GlobalStateMutex.Lock()
	defer orderManagement.GlobalStateMutex.Unlock()
	orderManagement.GlobalState.LocalID = elevID

	// Gjenopprett cab-orders hvis de fantes fra før
	oldState, exists := orderManagement.GlobalState.States[elevID]
	if exists {
		for f := 0; f < management.NumFloors; f++ {
			if oldState.CabRequests[f] {
				management.Elev.Orders[f][elevio.BT_Cab].OrderPlaced = true
				management.Elev.Orders[f][elevio.BT_Cab].Finished = false
				management.Elev.Orders[f][elevio.BT_Cab].ElevID = management.Elev.ID
			}
		}
	}

	// Registrer oss selv i GlobalState
	orderManagement.GlobalState.States[elevID] = orderManagement.ConvertElevatorToJSON(management.Elev)
}
