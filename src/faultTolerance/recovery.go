package faultTolerance

import (
	"fmt"
	"heislab/elevator/elevio"
	"heislab/hallRequestAssigner"
	"heislab/management"
	"heislab/orderManagement"
	"time"
)

// listens and merge global view received on startup. Returns true if received view
func RecoverOnStartup(gs *orderManagement.GlobalState, rx <-chan orderManagement.GlobalStateType) bool {
	elevID := management.Elev.ID
	timeout := time.After(1 * time.Second)
	var recovered *hallRequestAssigner.ElevatorStateJSON

	for {
		select {
		case globalState := <-rx:
			gs.Merge(globalState)

			if st, exists := globalState.States[elevID]; exists {
				tmp := st
				recovered = &tmp
			}
		case <-timeout:
			goto RECOVER
		}
	}

RECOVER:
	if recovered == nil {
		fmt.Println("No previous cab state found on startup, starting fresh")
		return false
	} else {
		if recovered.Floor >= 0 && recovered.Floor < management.NumFloors {
			management.Elev.Floor = recovered.Floor
			management.Elev.LastFloor = recovered.Floor
			elevio.SetFloorIndicator(recovered.Floor)
		}

		for floor := 0; floor < management.NumFloors && floor < len(recovered.CabRequests); floor++ {
			if recovered.CabRequests[floor] {
				management.Elev.Orders[floor][elevio.CabButton].OrderPlaced = true
				management.Elev.Orders[floor][elevio.CabButton].ElevID = elevID
				elevio.SetButtonLamp(elevio.CabButton, floor, true)
			}
		}
	}

	// Register local elevator state in global state after recovery.
	gs.SetElevatorGlobalState(elevID, orderManagement.ConvertElevatorToJSON(management.Elev))
	return true
}
