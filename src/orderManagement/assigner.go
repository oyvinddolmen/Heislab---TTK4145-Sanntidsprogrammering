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

	hallRequests := append([][2]bool(nil), globalStateCopy.HallRequests...)

	filtered := make(map[string]hallRequestAssigner.ElevatorStateJSON)
	for elevID, s := range globalStateCopy.States {
		if s.Behavior != "offline" {
			filtered[elevID] = s
		}
	}

	assignments, err := hallRequestAssigner.AssignHallRequests(hallRequests, filtered)
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
