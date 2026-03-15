package orderManagement

import (
	"fmt"
	"heislab/management"
	"heislab/state"
	"encoding/json"
	"os/exec"
	"runtime"
)

type AssignerInput struct {
	HallRequests [][2]bool                    `json:"hallRequests"`
	States       map[string]state.ElevatorStateJSON `json:"states"`
}

// RunHallAssigner kopierer hall requests og online elevator states,
// kaller hallRequestAssigner, og oppdaterer lokalt heis-objekt
func RunHallAssignerAndApplyAssignments(elev *management.Elevator, globalState *state.GlobalState) {
	globalStateCopy := globalState.GetCopy()
	hallRequests := append([][2]bool(nil), globalStateCopy.HallRequests...) // TODO: Unødvendig?

	activeElevators := make(map[string]state.ElevatorStateJSON)
	for elevID, state := range globalStateCopy.States {
		if state.Behavior != "offline" && state.CanTakeOrders {
			activeElevators[elevID] = state
		}
	}

	assignments, err := AssignHallRequests(hallRequests, activeElevators)
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

// AssignHallRequests calls the external hall_request_assigner binary
// and returns hall requests assigned per elevator.
func AssignHallRequests(
	hallRequests [][2]bool,
	states map[string]state.ElevatorStateJSON,
) (map[string][][2]bool, error) {

	input := AssignerInput{
		HallRequests: hallRequests,
		States:       states,
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal failed: %w", err)
	}
	jsonStr := string(jsonBytes)
	//fmt.Println("JSON sendt til hall_request_assigner:", jsonStr)

	assignerPath := ""
	switch runtime.GOOS {
	case "windows":
		assignerPath = "./hallRequestAssigner/hall_request_assigner.exe"
	case "darwin":
		assignerPath = "./hallRequestAssigner/hall_request_assigner_mac"
	default:
		assignerPath = "./hallRequestAssigner/hall_request_assigner"
	}

	// Run binary file
	cmd := exec.Command(assignerPath, "--input", jsonStr)

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf(
			"hall_request_assigner failed: %w\n%s",
			err,
			string(outputBytes),
		)
	}

	output := make(map[string][][2]bool)
	err = json.Unmarshal(outputBytes, &output)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed: %w\nOutput: %s", err, string(outputBytes))
	}

	return output, nil
}
