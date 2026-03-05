package orderManagement

import (
	"fmt"
	"heislab/hallRequestAssigner"
	"heislab/management"
)

// RunHallAssigner kopierer hall requests og online elevator states,
// kaller hallRequestAssigner, og oppdaterer lokalt heis-objekt
func RunHallAssigner(gs *GlobalState) {
	// Lås globalState og kopier hallRequests + online elevator states
	gs.mu.Lock()
	hallRequests := append([][2]bool(nil), gs.globalState.HallRequests...)

	filtered := make(map[string]hallRequestAssigner.ElevatorStateJSON)
	for id, s := range gs.globalState.States {
		if s.Behavior != "offline" {
			filtered[id] = s
		}
	}
	gs.mu.Unlock()

	gs.Print() // debug - printer GlobalState
	assignments, err := hallRequestAssigner.AssignHallRequests(hallRequests, filtered)
	if err != nil {
		fmt.Println("assigner failed: %w", err)
	}
	applyAssignments(assignments)
}

// applyAssignments oppdaterer lokal heis med tildelte hall-orders
func applyAssignments(assignments map[string][][2]bool) {
	elevID := management.Elev.ID
	assigned, exists := assignments[elevID]
	if !exists {
		fmt.Println("assignments[elevID] finnes ikke!!!")
		return
	}

	fmt.Println("assigned: ", assigned)

	for floor := 0; floor < management.NumFloors; floor++ {
		for btn := 0; btn < management.CabButton; btn++ { // only hall buttons
			if assigned[floor][btn] {
				management.Elev.Orders[floor][btn].OrderPlaced = true
				management.Elev.Orders[floor][btn].ElevID = management.Elev.ID
			} else{
				management.Elev.Orders[floor][btn].OrderPlaced = false
			}
		}
	}

	UpdateCurrentOrder()
	fmt.Println("current order:", management.Elev.CurrentOrder.Floor)
}