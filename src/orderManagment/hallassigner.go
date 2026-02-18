package orderManagment


import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type AssignerInput struct {
	HallRequests [][2]bool                `json:"hallRequests"`
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

	assignerPath := "orderManagment/hall_request_assigner"
	cmd := exec.Command(assignerPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe failed: %w", err)
	}

	go func() {
		defer stdin.Close()
		stdin.Write(jsonBytes)
	}()

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
		return nil, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	return output, nil
}

