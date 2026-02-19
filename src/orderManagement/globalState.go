package orderManagement

import (
	"heislab/management"
	"strconv"
	"sync"
	"heislab/hallRequestAssigner"
)

type GlobalStateType struct {
	HallRequests [][2]bool                    // [floor][0=up,1=down]
	States       map[string]hallRequestAssigner.ElevatorStateJSON // elevatorID -> state
}

var (
    GlobalState GlobalStateType
    GlobalStateMutex sync.Mutex
)

func InitGlobalState() {
	GlobalStateMutex.Lock()
	defer GlobalStateMutex.Unlock()

	GlobalState.HallRequests = make([][2]bool, management.NumFloors)
	GlobalState.States = make(map[string]hallRequestAssigner.ElevatorStateJSON)
}

// Convert elevator to JSON elevator state
func ConvertElevatorToJSON(e management.Elevator) hallRequestAssigner.ElevatorStateJSON {

	cabRequests := make([]bool, management.NumFloors)
	for f := 0; f < management.NumFloors; f++ {
		cabRequests[f] = e.Orders[f][2].OrderPlaced // 2 = Cab button
	}

	return hallRequestAssigner.ElevatorStateJSON{
		Behavior:    convertState(e.State),
		Floor:       e.Floor,
		Direction:   convertDirection(e.MoveDir),
		CabRequests: cabRequests,
	}
}

func convertState(s management.State) string {
	switch s {
	case management.IDLE:
		return "idle"
	case management.MOVING:
		return "moving"
	case management.DOOROPEN:
		return "doorOpen"
	case management.OFFLINE:
		return "offline"
	default:
		return "idle"
	}
}

func convertDirection(d management.Direction) string {
	switch d {
	case management.Dir_Up:
		return "up"
	case management.Dir_Down:
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
	GlobalStateMutex.Lock()
	defer GlobalStateMutex.Unlock()

	id := strconv.Itoa(management.Elev.ID)
	GlobalState.States[id] = ConvertElevatorToJSON(management.Elev)
}

// Update hall requests from Elev.Orders
func UpdateHallRequests() {
	GlobalStateMutex.Lock()
	defer GlobalStateMutex.Unlock()

	for f := 0; f < management.NumFloors; f++ {
		GlobalState.HallRequests[f][0] = management.Elev.Orders[f][0].OrderPlaced // HallUp
		GlobalState.HallRequests[f][1] = management.Elev.Orders[f][1].OrderPlaced // HallDown
	}
}

// Merge received remote elevator state
func MergeRemoteElevator(id string, e management.Elevator) {
	GlobalStateMutex.Lock()
	defer GlobalStateMutex.Unlock()

	GlobalState.States[id] = ConvertElevatorToJSON(e)
}
