package network

import (
	"heislab/management"
	"heislab/state"
	"heislab/elevator/elevIO"
	"time"
	"fmt"
)

// Listens and merges global state received on startup. Returns true if received state.
func RecoverOnStartup(
	elev *management.Elevator,
	globalState *state.GlobalState,
	incomingGlobalStateChannel <-chan state.GlobalStateData,
) bool {
	elevID := globalState.GetLocalID()
	timeout := time.After(1 * time.Second)
	var recoveredElevatorState *state.ElevatorStateJSON

	for {
		select {
		case newGlobalState := <-incomingGlobalStateChannel:
			globalState.Merge(newGlobalState)

			if elevState, exists := newGlobalState.States[elevID]; exists {
				recoveredElevatorState = new(state.ElevatorStateJSON)
				*recoveredElevatorState = elevState
			}
		case <-timeout:
			goto RECOVER
		}
	}

RECOVER:
	if recoveredElevatorState == nil {
		fmt.Println("No previous cab state found on startup, starting fresh")
		return false
	} else {
		if recoveredElevatorState.Floor >= 0 && recoveredElevatorState.Floor < management.NumFloors {
			elev.SetFloor(recoveredElevatorState.Floor)
			elev.SetLastFloor(recoveredElevatorState.Floor)
			elevIO.SetFloorIndicator(recoveredElevatorState.Floor)
		}

		for floor := 0; floor < management.NumFloors && floor < len(recoveredElevatorState.CabOrders); floor++ {
			if recoveredElevatorState.CabOrders[floor] {
				elev.SetOrderActiveStatus(floor, int(elevIO.CabButton), true)
				elevIO.SetButtonLamp(elevIO.CabButton, floor, true)
			}
		}
	}

	// Register local elevator state in global state after recovery.
	globalState.SetElevatorGlobalState(elevID, state.ConvertElevatorToJSON(elev))
	return true
}
