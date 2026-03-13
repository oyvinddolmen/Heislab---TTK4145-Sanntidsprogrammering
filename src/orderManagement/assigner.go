package orderManagement

import (
	"fmt"
	"heislab/hallRequestAssigner"
	"heislab/management"
	"heislab/state"
)

// RunHallAssigner kopierer hall requests og online elevator states,
// kaller hallRequestAssigner, og oppdaterer lokalt heis-objekt
func RunHallAssignerAndApplyAssignments(elev *management.Elevator, gs *state.GlobalState) {
	gsCopy := gs.GetCopy()

	hallRequests := append([][2]bool(nil), gsCopy.HallRequests...)

	filtered := make(map[string]hallRequestAssigner.ElevatorStateJSON)
	for id, s := range gsCopy.States {
		if s.Behavior != "offline" && s.CanTakeOrders {
			filtered[id] = s
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
		for btn := 0; btn < management.CabButton; btn++ { // only hall buttons
			if assigned[floor][btn] {
				elev.Orders[floor][btn].OrderPlaced = true
			} else {
				elev.Orders[floor][btn].OrderPlaced = false
			}
		}
	}
}
