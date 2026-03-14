package orderManagement

import (
	"fmt"
	"heislab/hallRequestAssigner"
	"heislab/management"
	"heislab/state"
)

// RunHallAssigner kopierer hall requests og online elevator states,
// kaller hallRequestAssigner, og oppdaterer lokalt heis-objekt
func RunHallAssignerAndApplyAssignments(elev *management.Elevator, globalState *state.GlobalState) {
	globalStateCopy := globalState.GetCopy()
	hallRequests := append([][2]bool(nil), globalStateCopy.HallRequests...) // TODO: Unødvendig?

	activeElevators := make(map[string]hallRequestAssigner.ElevatorStateJSON)
	for elevID, state := range globalStateCopy.States {
		if state.Behavior != "offline" && state.CanTakeOrders {
			activeElevators[elevID] = state
		}
	}

	assignments, err := hallRequestAssigner.AssignHallRequests(hallRequests, activeElevators)
	if err != nil {
		fmt.Println("assigner failed: %w", err)
	}
	applyAssignments(elev, assignments)
}

// applyAssignments oppdaterer lokal heis med tildelte hall-orders
func applyAssignments(elev *management.Elevator, assignments map[string][][2]bool) {
	elevID := elev.GetElevID()
	assigned, exists := assignments[elevID]
	if !exists {
		fmt.Println("assignments[elevID] finnes ikke!!!")
		return
	}

	//fmt.Println("assigned: ", assigned)

	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < management.CabButton; button++ { // only hall buttons
			if assigned[floor][button] {
				elev.Orders[floor][button].OrderPlaced = true
			} else {
				elev.Orders[floor][button].OrderPlaced = false
			}
		}
	}
}
