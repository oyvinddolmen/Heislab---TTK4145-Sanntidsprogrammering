package hallRequestAssigner

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

type ElevatorStateJSON struct {
	Behavior      string `json:"behaviour"` // idle, moving, doorOpen, offline
	Floor         int    `json:"floor"`
	Direction     string `json:"direction"`
	CabRequests   []bool `json:"cabRequests"`
	CanTakeOrders bool   `json:"canTakeOrders"`
}

type AssignerInput struct {
	HallRequests [][2]bool                    `json:"hallRequests"`
	States       map[string]ElevatorStateJSON `json:"states"`
}

// AssignHallRequests calls the external hall_request_assigner binary
// and returns hall requests assigned per elevator.
func AssignHallRequests(
	hallRequests [][2]bool,
	states map[string]ElevatorStateJSON,
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

	//fmt.Println("AHR ferdig kjørt. Output:", output)
	return output, nil
}
