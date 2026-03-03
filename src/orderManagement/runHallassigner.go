package orderManagement

import (
	"fmt"
	"heislab/management"
	"heislab/hallRequestAssigner"
)

func RunHallAssigner() error {
	GlobalStateMutex.Lock()
	// Copy HallRequests
	hallRequests := make([][2]bool, len(GlobalState.HallRequests))
	copy(hallRequests, GlobalState.HallRequests)

	// CopyStates (only the online elevators)
	filtered := make(map[string]hallRequestAssigner.ElevatorStateJSON)
	for id, s := range GlobalState.States {
		if s.Behavior != "offline" {
			filtered[id] = s
		}
	}
	GlobalStateMutex.Unlock()
	PrintGlobalState()
	
	assignments, err := hallRequestAssigner.AssignHallRequests(hallRequests, filtered)
	if err != nil {
		return fmt.Errorf("assigner failed: %w", err)
	}

	applyAssignments(assignments)
	return nil
}

func applyAssignments(assignments map[string][][2]bool) {
	localID := management.Elev.IP

	assigned, exists := assignments[localID]
	if !exists {
		return
	}
	fmt.Println("assigned: ",assigned)
	for floor := 0; floor < management.NumFloors; floor++ {
		for btn := 0; btn < management.CabButton; btn++ { // only hall buttons
			if assigned[floor][btn] {
				management.Elev.Orders[floor][btn].OrderPlaced = true
				management.Elev.Orders[floor][btn].ElevIP = management.Elev.IP
			}
		}
	}

	UpdateCurrentOrder()
}
