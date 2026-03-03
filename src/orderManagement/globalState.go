package orderManagement

import (
	"fmt"
	"heislab/hallRequestAssigner"
	"heislab/management"
	"sync"
)

type GlobalStateType struct {
	HallRequests        [][2]bool                                        // [floor][0=up,1=down]
	HallRequestsVersion [][2]int                                         //incremented by one when matching hallRequest is updated
	States              map[string]hallRequestAssigner.ElevatorStateJSON // elevatorID -> state
	LocalID             string
}

var (
	GlobalState      GlobalStateType
	GlobalStateMutex sync.Mutex
)

func InitGlobalState(elevID string) {
	GlobalStateMutex.Lock()
	defer GlobalStateMutex.Unlock()

	GlobalState.HallRequests = make([][2]bool, management.NumFloors)
	GlobalState.HallRequestsVersion = make([][2]int, management.NumFloors)
	GlobalState.States = make(map[string]hallRequestAssigner.ElevatorStateJSON)
	GlobalState.LocalID = elevID

	GlobalState.States[management.Elev.ID] = ConvertElevatorToJSON(management.Elev)
	GlobalState.States[management.Elev.ID] = ConvertElevatorToJSON(management.Elev)
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

func convertDirection(direction management.Direction) string {
	switch direction {
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

	GlobalState.States[management.Elev.ID] = ConvertElevatorToJSON(management.Elev)
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
	for f := 0; f < len(GlobalState.HallRequestsVersion); f++ {
		up := GlobalState.HallRequestsVersion[f][0]
		down := GlobalState.HallRequestsVersion[f][1]
		fmt.Printf("  Floor %d:  Up Version = %d Down Version = %d\n", f, up, down)
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

func MergeGlobalState(globalState GlobalStateType) {
	GlobalStateMutex.Lock()
	defer GlobalStateMutex.Unlock()

	localID := management.Elev.ID
	senderID := globalState.LocalID

	if senderID != localID {
		if st, exists := globalState.States[senderID]; exists {
			GlobalState.States[senderID] = st
		}
	}

	chooseLatestHallRequestVersions(globalState)
}

func chooseLatestHallRequestVersions(gs GlobalStateType) {
	for f := 0; f < management.NumFloors; f++ {
		for b := 0; b < 2; b++ {

			localVersion := GlobalState.HallRequestsVersion[f][b]
			remoteVersion := gs.HallRequestsVersion[f][b]

			switch {
			case remoteVersion > localVersion:
				// Remote er nyere → overta verdi og versjon
				GlobalState.HallRequests[f][b] = gs.HallRequests[f][b]
				GlobalState.HallRequestsVersion[f][b] = remoteVersion

			case remoteVersion == localVersion:
				// Samme versjon → true vinner
				if gs.HallRequests[f][b] {
					GlobalState.HallRequests[f][b] = true
				}
			}
		}
	}
}

func IncremtHallRequestVersion(btnPress management.Order) {
	GlobalState.HallRequestsVersion[btnPress.Floor][btnPress.ButtonType]++
	fmt.Print("+ på Version, ordre LAGT TIL")
}

// Set dead elevator behavior to "offline" and redistribute orders
func SetElevatorToOffline(deadID string) {
	GlobalStateMutex.Lock()
	defer GlobalStateMutex.Unlock()

	state, exists := GlobalState.States[deadID]
	if !exists {
		GlobalStateMutex.Unlock()
		return
	}

	state.Behavior = "offline"
	GlobalState.States[deadID] = state
}
