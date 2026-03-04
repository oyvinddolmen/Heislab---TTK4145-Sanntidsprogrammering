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

		case globalState := <-rx:
			orderManagement.GlobalStateMutex.Lock()
			orderManagement.GlobalState = globalState
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
		for floor := 0; floor < management.NumFloors; floor++ {
			if oldState.CabRequests[floor] {
				management.Elev.Orders[floor][elevio.CabButton].OrderPlaced = true
				management.Elev.Orders[floor][elevio.CabButton].Finished = false
				management.Elev.Orders[floor][elevio.CabButton].ElevID = management.Elev.ID
			}
		}
	}

	// Registrer oss selv i GlobalState
	orderManagement.GlobalState.States[elevID] = orderManagement.ConvertElevatorToJSON(management.Elev)
}
