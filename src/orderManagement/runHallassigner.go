package orderManagement

import (
	"fmt"
	"heislab/management"
	"strconv"
)

func RunHallAssigner() error {

	mutex.Lock()
	// Copy HallRequests
	hallRequests := make([][2]bool, len(globalState.HallRequests))
	copy(hallRequests, globalState.HallRequests)

	// CopyStates (only the online elevators)
	filtered := make(map[string]ElevatorStateJSON)
	for id, s := range globalState.States {
		if s.Behavior != "offline" {
			filtered[id] = s
		}
	}
	mutex.Unlock()

	assignments, err := AssignHallRequests(hallRequests, filtered)
	if err != nil {
		return fmt.Errorf("assigner failed: %w", err)
	}

	applyAssignments(assignments)
	return nil
}

func applyAssignments(assignments map[string][][2]bool) {

	mutex.Lock()
	defer mutex.Unlock()

	localID := strconv.Itoa(management.Elev.ID)

	assigned, exists := assignments[localID]
	if !exists {
		return
	}

	for floor := 0; floor < management.NumFloors; floor++ {
		for btn := 0; btn < 2; btn++ { // only hall buttons
			if assigned[floor][btn] {
				management.Elev.Orders[floor][btn].OrderPlaced = true
				management.Elev.Orders[floor][btn].ElevID = management.Elev.ID
			}
		}
	}

	UpdateCurrentOrder()
}
