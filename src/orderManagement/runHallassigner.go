package orderManagement

import (
	"fmt"
	"heislab/management"
	"heislab/hallRequestAssigner"
)

// RunHallAssigner kopierer hall requests og online elevator states,
// kaller hallRequestAssigner, og oppdaterer lokalt heis-objekt
func RunHallAssigner(gs *GlobalState) error {

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

	gs.Print() // debug

	assignments, err := hallRequestAssigner.AssignHallRequests(hallRequests, filtered)
	if err != nil {
		return fmt.Errorf("assigner failed: %w", err)
	}

	applyAssignments(gs, assignments)
	return nil
}

// applyAssignments oppdaterer lokal heis med tildelte hall-orders
func applyAssignments(gs *GlobalState, assignments map[string][][2]bool) {
	elevID := management.Elev.ID
	assigned, exists := assignments[elevID]
	if !exists {
		return
	}

	fmt.Println("assigned: ", assigned)

	for floor := 0; floor < management.NumFloors; floor++ {
		for btn := 0; btn < management.CabButton; btn++ { // only hall buttons
			if assigned[floor][btn] {
				management.Elev.Orders[floor][btn].OrderPlaced = true
				management.Elev.Orders[floor][btn].ElevID = management.Elev.ID
			}
		}
	}

	// Oppdater globalState med nye lokale hall requests
	gs.mu.Lock()
	for floor := 0; floor < management.NumFloors; floor++ {
		for btn := 0; btn < management.CabButton; btn++ {
			if assigned[floor][btn] {
				gs.globalState.HallRequests[floor][btn] = true
			}
		}
	}
	gs.mu.Unlock()

	UpdateCurrentOrder(gs)
}