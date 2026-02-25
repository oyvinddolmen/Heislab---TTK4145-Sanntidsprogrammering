package orderManagement

import (
	"heislab/management"
	"sync"
	"heislab/hallRequestAssigner"
)

type GlobalStateType struct {
	HallRequests         [][2]bool // [floor][0=up,1=down]
	HallRequestsAssigned [][2]string
	States               map[string]hallRequestAssigner.ElevatorStateJSON // elevatorID -> state
	LocalIP		 		 string
}

var (
    GlobalState GlobalStateType
    GlobalStateMutex sync.Mutex
)

func InitGlobalState(localIP string) {
	GlobalStateMutex.Lock()
	defer GlobalStateMutex.Unlock()

	GlobalState.HallRequests = make([][2]bool, management.NumFloors)
	GlobalState.HallRequestsAssigned = make([][2]string, management.NumFloors)
	GlobalState.States = make(map[string]hallRequestAssigner.ElevatorStateJSON)
	GlobalState.LocalIP = localIP

	GlobalState.States[management.Elev.IP] = ConvertElevatorToJSON(management.Elev)
}

// Convert elevator to JSON elevator state
func ConvertElevatorToJSON(e management.Elevator) hallRequestAssigner.ElevatorStateJSON {

	cabRequests := make([]bool, management.NumFloors)
	for f := 0; f < management.NumFloors; f++ {
		cabRequests[f] = e.Orders[f][2].OrderPlaced // 2 = Cab button
	}

	return hallRequestAssigner.ElevatorStateJSON{
		Behavior:    convertState(e.State),
		Floor:       e.LastFloor,
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
	case management.INIT:
		return "moving"
	case management.STOP:
		return "STOP"
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

	GlobalState.States[management.Elev.IP] = ConvertElevatorToJSON(management.Elev)
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
