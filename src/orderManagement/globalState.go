package orderManagement

import (
	"heislab/management"
	"strconv"
	"sync"
	"heislab/hallRequestAssigner"
	"fmt"
)

type GlobalStateType struct {
	HallRequests         [][2]bool // [floor][0=up,1=down]
	HallRequestsAssigned [][2]string
	States               map[string]hallRequestAssigner.ElevatorStateJSON // elevatorID -> state
}

var (
    GlobalState GlobalStateType
    GlobalStateMutex sync.Mutex
)

func InitGlobalState() {
	GlobalStateMutex.Lock()
	defer GlobalStateMutex.Unlock()

	GlobalState.HallRequests = make([][2]bool, management.NumFloors)
	GlobalState.HallRequestsAssigned = make([][2]string, management.NumFloors)
	GlobalState.States = make(map[string]hallRequestAssigner.ElevatorStateJSON)

	id := strconv.Itoa(management.Elev.ID)
	GlobalState.States[id] = ConvertElevatorToJSON(management.Elev)
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
	GlobalStateMutex.Lock()
	defer GlobalStateMutex.Unlock()

	id := strconv.Itoa(management.Elev.ID)
	GlobalState.States[id] = ConvertElevatorToJSON(management.Elev)

	for f := 0; f < management.NumFloors; f++ {
		GlobalState.HallRequests[f][0] = management.Elev.Orders[f][0].OrderPlaced // HallUp
		GlobalState.HallRequests[f][1] = management.Elev.Orders[f][1].OrderPlaced // HallDown
	}
}

func PrintGlobalState() {

	GlobalStateMutex.Lock()
	defer GlobalStateMutex.Unlock()

	fmt.Println("====================================")
	fmt.Println("           GLOBAL STATE             ")
	fmt.Println("====================================")

	// ---- Hall Requests ----
	fmt.Println("\nHallRequests:")
	for f := 0; f < len(GlobalState.HallRequests); f++ {
		up := GlobalState.HallRequests[f][0]
		down := GlobalState.HallRequests[f][1]
		fmt.Printf("  Floor %d: Up=%v Down=%v\n", f, up, down)
	}

	// ---- Hall Assignments ----
	fmt.Println("\nHallRequestsAssigned:")
	for f := 0; f < len(GlobalState.HallRequestsAssigned); f++ {
		up := GlobalState.HallRequestsAssigned[f][0]
		down := GlobalState.HallRequestsAssigned[f][1]
		fmt.Printf("  Floor %d: Up=%s Down=%s\n", f, up, down)
	}

	// ---- Elevator States ----
	fmt.Println("\nElevator States:")
	for id, state := range GlobalState.States {

		fmt.Printf("\n  Elevator %s\n", id)
		fmt.Printf("    Behaviour : %s\n", state.Behavior)
		fmt.Printf("    Floor     : %d\n", state.Floor)
		fmt.Printf("    Direction : %s\n", state.Direction)

		fmt.Printf("    CabRequests: ")
		for f := 0; f < len(state.CabRequests); f++ {
			if state.CabRequests[f] {
				fmt.Printf("[%d] ", f)
			}
		}
		fmt.Println()
	}

	fmt.Println("\n====================================")
}