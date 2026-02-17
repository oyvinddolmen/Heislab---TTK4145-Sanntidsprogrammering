package orderManagment

import (
	"fmt"
	"heislab/managment"
	"strconv"
	"sync"
)

type GlobalState struct {
	HallRequests [][2]bool                // [floor][0=up,1=down]
	States       map[string]ElevatorStateJSON // elevatorID -> state
}

var globalState GlobalState
var mutex sync.Mutex


func InitGlobalState() {
	mutex.Lock()
	defer mutex.Unlock()

	globalState.HallRequests = make([][2]bool, managment.NumFloors)
	globalState.States = make(map[string]ElevatorStateJSON)
}


// Convert elevator to JSON elevator state
func convertElevatorToJSON(e managment.Elevator) ElevatorStateJSON {

	cabRequests := make([]bool, managment.NumFloors)
	for f := 0; f < managment.NumFloors; f++ {
		cabRequests[f] = e.Orders[f][2].OrderPlaced // 2 = Cab button
	}

	return ElevatorStateJSON{
		Behavior:    convertState(e.State),
		Floor:       e.Floor,
		Direction:   convertDirection(e.MoveDir),
		CabRequests: cabRequests,
	}
}

func convertState(s managment.State) string {
	switch s {
	case managment.IDLE:
		return "idle"
	case managment.EXECUTING:
		return "moving"
	case managment.DOOROPEN:
		return "doorOpen"
	default:
		return "idle"
	}
}

func convertDirection(d managment.Direction) string {
	switch d {
	case managment.Dir_Up:
		return "up"
	case managment.Dir_Down:
		return "down"
	default:
		return "stop"
	}
}

func UpdateLocalGlobalState() {
    UpdateLocalElevator()
    UpdateHallRequests()
}

// Update local elevator state in globalState
func UpdateLocalElevator() {
	mutex.Lock()
	defer mutex.Unlock()

	id := strconv.Itoa(managment.Elev.ID)
	globalState.States[id] = convertElevatorToJSON(managment.Elev)
}

// Update hall requests from Elev.Orders
func UpdateHallRequests() {
	mutex.Lock()
	defer mutex.Unlock()

	for f := 0; f < managment.NumFloors; f++ {
		globalState.HallRequests[f][0] = managment.Elev.Orders[f][0].OrderPlaced // HallUp
		globalState.HallRequests[f][1] = managment.Elev.Orders[f][1].OrderPlaced // HallDown
	}
}

// Merge received remote elevator state
func MergeRemoteElevator(id string, e managment.Elevator) {
	mutex.Lock()
	defer mutex.Unlock()

	globalState.States[id] = convertElevatorToJSON(e)
}


func RunHallAssigner() error {

	mutex.Lock()
	// Copy HallRequests
	hallRequests := make([][2]bool, len(globalState.HallRequests))
	copy(hallRequests, globalState.HallRequests)

	// CopyStates
	states := make(map[string]ElevatorStateJSON, len(globalState.States))
	for k, v := range globalState.States {
		states[k] = v
	}
	mutex.Unlock()

	assignments, err := AssignHallRequests(hallRequests, states)
	if err != nil {
		return fmt.Errorf("assigner failed: %w", err)
	}

	applyAssignments(assignments)
	return nil
}


func applyAssignments(assignments map[string][][2]bool) {

	mutex.Lock()
	defer mutex.Unlock()

	localID := strconv.Itoa(managment.Elev.ID)

	assigned, exists := assignments[localID]
	if !exists {
		return
	}

	for floor := 0; floor < managment.NumFloors; floor++ {
		for btn := 0; btn < 2; btn++ { // only hall buttons
			if assigned[floor][btn] {
				managment.Elev.Orders[floor][btn].OrderPlaced = true
				managment.Elev.Orders[floor][btn].Status = managment.Elev.ID
			}
		}
	}
}
