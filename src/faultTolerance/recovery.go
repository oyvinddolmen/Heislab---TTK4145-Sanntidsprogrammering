package faultTolerance

import (
	"fmt"
	"heislab/elevio"
	"heislab/hallRequestAssigner"
	"heislab/management"
	"heislab/orderManagement"
	"time"
)

func RecoverOnStartup(gs *orderManagement.GlobalState, rx <-chan orderManagement.GlobalStateType) {
	elevID := management.Elev.ID
	timeout := time.After(1 * time.Second)
	var recovered *hallRequestAssigner.ElevatorStateJSON

	// Listen for existing global states for a short window.
	// We look specifically for our own previous state in others' world views.
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
		return
	} else {
		for floor := 0; floor < management.NumFloors && floor < len(recovered.CabRequests); floor++ {
			if recovered.CabRequests[floor] {
				management.Elev.Orders[floor][elevio.CabButton].OrderPlaced = true
				management.Elev.Orders[floor][elevio.CabButton].ElevID = elevID
				elevio.SetButtonLamp(elevio.CabButton, floor, true)
			}
		}
	}

	// Register local elevator state in global state after recovery.
	gs.SetElevatorState(elevID, orderManagement.ConvertElevatorToJSON(management.Elev))
}
