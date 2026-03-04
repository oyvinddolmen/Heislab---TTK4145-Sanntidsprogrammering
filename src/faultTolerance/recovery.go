package faultTolerance

import (
	"fmt"
	"heislab/elevio"
	"heislab/management"
	"heislab/orderManagement"
	"time"
)

func RecoverOnStartup(gs *orderManagement.GlobalState, rx <-chan orderManagement.GlobalStateType) {
	timeout := time.After(1 * time.Second)

	// Vent på eksisterende GlobalState
	for {
		select {
		case globalState := <-rx:
			gs.Merge(globalState) // merge innkommende state
			goto RECOVER
		case <-timeout:
			fmt.Println("No GlobalState received on startup, starting fresh")
			goto RECOVER
		}
	}

RECOVER:

	elevID := management.Elev.ID

	// Gjenopprett cab-orders hvis de fantes fra før
	if oldState, exists := gs.GetElevatorState(elevID); exists {
		for floor := 0; floor < management.NumFloors; floor++ {
			if oldState.CabRequests[floor] {
				management.Elev.Orders[floor][elevio.CabButton].OrderPlaced = true
				//management.Elev.Orders[floor][elevio.CabButton].Finished = false
				management.Elev.Orders[floor][elevio.CabButton].ElevID = elevID
			}
		}
	}

	// Registrer oss selv i GlobalState via metode
	gs.SetElevatorState(elevID, orderManagement.ConvertElevatorToJSON(management.Elev))
}