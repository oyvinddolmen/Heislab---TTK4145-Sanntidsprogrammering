package orderManagement

import (
	"encoding/json"
	"fmt"
	"heislab/management"
	"heislab/state"
	"os/exec"
	"runtime"
)

type AssignerInput struct {
	HallOrders     [][management.NumHallButtonTypes]bool `json:"hallRequests"`
	ElevatorStates map[string]state.ElevatorStateJSON    `json:"states"`
}

// Assigns hall orders and applies assigned hall orders to the local elevator.
func RunHallAssignerAndApplyAssignments(elev *management.Elevator, globalState *state.GlobalState) {
	globalStateCopy := globalState.GetCopy()

	for id, s := range globalStateCopy.States {
		fmt.Printf(
			"before assigner: id=%s floor=%d behavior=%s dir=%s canTakeOrders=%v\n",
			id, s.Floor, s.Behavior, s.Direction, s.CanTakeOrders,
		)
	}

	activeElevatorStates := make(map[string]state.ElevatorStateJSON)
	for elevID, elevState := range globalStateCopy.States {
		if elevState.Behavior != "offline" && elevState.CanTakeOrders {
			activeElevatorStates[elevID] = elevState
		}
	}

	assignments, err := AssignHallOrders(globalStateCopy.HallOrders, activeElevatorStates)
	if err != nil {
		fmt.Println("assigner failed: %w", err)
	}
	applyAssignments(elev, assignments)
}

// Applies assigned hall orders to the local elevator.
func applyAssignments(elev *management.Elevator, assignments map[string][][management.NumHallButtonTypes]bool) {
	elevID := elev.GetID()
	assigned, exists := assignments[elevID]
	if !exists {
		fmt.Println("assignments[elevID] finnes ikke!!!")
		return
	}

	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < management.NumHallButtonTypes; button++ {
			if assigned[floor][button] {
				elev.SetOrderActiveStatus(floor, button, true)
			} else {
				elev.SetOrderActiveStatus(floor, button, false)
			}
		}
	}
}

// Finds and returns hall orders assigned per elevator.
func AssignHallOrders(
	hallOrders [][management.NumHallButtonTypes]bool,
	elevatorStates map[string]state.ElevatorStateJSON,
) (map[string][][management.NumHallButtonTypes]bool, error) {

	input := AssignerInput{
		HallOrders:     hallOrders,
		ElevatorStates: elevatorStates,
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal failed: %w", err)
	}
	jsonString := string(jsonBytes)

	assignerPath := ""
	switch runtime.GOOS {
	case "windows":
		assignerPath = "./orderManagement/hallOrderAssigner/hall_request_assigner.exe"
	case "darwin":
		assignerPath = "./orderManagement/hallOrderAssigner/hall_request_assigner_mac"
	default:
		assignerPath = "./orderManagement/hallOrderAssigner/hall_request_assigner"
	}

	// Run binary file
	cmd := exec.Command(assignerPath, "--input", jsonString)

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf(
			"hall_request_assigner failed: %w\n%s",
			err,
			string(outputBytes),
		)
	}

	assignments := make(map[string][][management.NumHallButtonTypes]bool)
	err = json.Unmarshal(outputBytes, &assignments)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed: %w\nOutput: %s", err, string(outputBytes))
	}

	return assignments, nil
}
