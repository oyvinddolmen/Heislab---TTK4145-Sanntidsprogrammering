package orderManagement

import (
	"fmt"
	"heislab/hallRequestAssigner"
	"heislab/management"
	"sync"
)

type GlobalStateType struct {
	HallRequests        [][2]bool                                        // [floor][0=up,1=down]
	HallRequestsVersion [][2]int                                         // incremented by one when matching hallRequest is updated
	States              map[string]hallRequestAssigner.ElevatorStateJSON // elevatorID -> state
	LocalID             string
}

// GlobalState kapsler globalState + mutex
type GlobalState struct {
	mu          sync.Mutex
	globalState GlobalStateType
}

// Constructor
func NewGlobalState(elevID string) *GlobalState {
	gs := &GlobalState{}

	gs.globalState.HallRequests = make([][2]bool, management.NumFloors)
	gs.globalState.HallRequestsVersion = make([][2]int, management.NumFloors)
	gs.globalState.States = make(map[string]hallRequestAssigner.ElevatorStateJSON)
	gs.globalState.LocalID = elevID

	// initial local elevator state
	gs.globalState.States[management.Elev.ID] = ConvertElevatorToJSON(management.Elev)

	return gs
}

// -------------------- State Conversion --------------------

func ConvertElevatorToJSON(e management.Elevator) hallRequestAssigner.ElevatorStateJSON {
	cabRequests := make([]bool, management.NumFloors)
	for floor := 0; floor < management.NumFloors; floor++ {
		cabRequests[floor] = e.Orders[floor][management.CabButton].OrderPlaced
	}

	return hallRequestAssigner.ElevatorStateJSON{
		Behavior:    convertState(e.State),
		Floor:       e.LastFloor,
		Direction:   convertDirection(e.MoveDir),
		CabRequests: cabRequests,
	}
}

func convertState(state management.State) string {
	switch state {
	case management.ElevIdle:
		return "idle"
	case management.ElevMoving:
		return "moving"
	case management.ElevInit:
		return "moving"
	case management.ElevStop:
		return "doorOpen"
	case management.ElevOffline:
		return "offline"
	default:
		return "idle"
	}
}

func convertDirection(direction management.Direction) string {
	switch direction {
	case management.DirUp:
		return "up"
	case management.DirDown:
		return "down"
	default:
		return "stop"
	}
}

// -------------------- Public Methods --------------------

func (gs *GlobalState) UpdateLocalGlobalState() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.globalState.States[management.Elev.ID] = ConvertElevatorToJSON(management.Elev)
}

func (gs *GlobalState) AddHallRequest(order management.Order) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.globalState.HallRequests[order.Floor][order.ButtonType] = true
}

func (gs *GlobalState) IncrementHallRequestVersion(order management.Order) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.globalState.HallRequestsVersion[order.Floor][order.ButtonType]++
}

func (gs *GlobalState) SetElevatorToOffline(deadID string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	elevState, exists := gs.globalState.States[deadID]
	if !exists {
		return
	}

	elevState.Behavior = "offline"
	gs.globalState.States[deadID] = elevState
}

// -------------------- Merge Logic --------------------

func (gs *GlobalState) Merge(remote GlobalStateType) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	localID := gs.globalState.LocalID
	senderID := remote.LocalID

	if senderID != localID {
		if st, exists := remote.States[senderID]; exists {
			gs.globalState.States[senderID] = st
		}
	}

	chooseLatestHallRequestVersions(&gs.globalState, remote)
}

func chooseLatestHallRequestVersions(local *GlobalStateType, remote GlobalStateType) {
	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < 2; button++ {
			localV := local.HallRequestsVersion[floor][button]
			remoteV := remote.HallRequestsVersion[floor][button]

			switch {
			case remoteV > localV:
				local.HallRequests[floor][button] = remote.HallRequests[floor][button]
				local.HallRequestsVersion[floor][button] = remoteV
			case remoteV == localV:
				if remote.HallRequests[floor][button] {
					local.HallRequests[floor][button] = true
				}
			}
		}
	}
}


// SetElevatorState setter en spesifikk elevator state i globalState
func (gs *GlobalState) SetElevatorState(elevID string, state hallRequestAssigner.ElevatorStateJSON) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.globalState.States[elevID] = state
}

// GetElevatorState henter state for en spesifikk heis
func (gs *GlobalState) GetElevatorState(elevID string) (hallRequestAssigner.ElevatorStateJSON, bool) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	state, ok := gs.globalState.States[elevID]
	return state, ok
}

// -------------------- World View Comparison --------------------
//TROR IKKE VI TRENGER DENNE. DEN SJEKKER BARE OM DET ER NOEN NYE OPPDATERINGER
/*
func (gs *GlobalState) NewWorldVie(remote GlobalStateType) bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// hall request changes
	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < 2; button++ {
			localV := gs.globalState.HallRequestsVersion[floor][button]
			remoteV := remote.HallRequestsVersion[floor][button]

			if remoteV > localV {
				return true
			}
			if remoteV == localV && remote.HallRequests[floor][button] && !gs.globalState.HallRequests[floor][button] {
				return true
			}
		}
	}

	// elevator state changes
	senderID := remote.LocalID
	if senderID == "" || senderID == management.Elev.ID {
		return false
	}

	remoteState, ok := remote.States[senderID]
	if !ok {
		return false
	}

	localState, exists := gs.globalState.States[senderID]
	if !exists {
		return true
	}

	if remoteState.Behavior != localState.Behavior ||
		remoteState.Floor != localState.Floor ||
		remoteState.Direction != localState.Direction {
		return true
	}

	if len(remoteState.CabRequests) != len(localState.CabRequests) {
		return true
	}
	for i := range remoteState.CabRequests {
		if remoteState.CabRequests[i] != localState.CabRequests[i] {
			return true
		}
	}

	return false
}
*/

// -------------------- Safe Getter --------------------

func (gs *GlobalState) GetCopy() GlobalStateType {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	copyState := gs.globalState

	// deep copy slices
	copyState.HallRequests = append([][2]bool(nil), gs.globalState.HallRequests...)
	copyState.HallRequestsVersion = append([][2]int(nil), gs.globalState.HallRequestsVersion...)

	newMap := make(map[string]hallRequestAssigner.ElevatorStateJSON)
	for k, v := range gs.globalState.States {
		newMap[k] = v
	}
	copyState.States = newMap

	return copyState
}

// -------------------- Debug --------------------

func (gs *GlobalState) Print() {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	fmt.Println("\n========== GLOBAL STATE ==========")
	fmt.Printf("LocalID: %s\n", gs.globalState.LocalID)

	// ---------------- Hall Requests ----------------
	fmt.Println("\nHall Requests:")
	for floor := 0; floor < len(gs.globalState.HallRequests); floor++ {
		up := gs.globalState.HallRequests[floor][0]
		down := gs.globalState.HallRequests[floor][1]
		upV := gs.globalState.HallRequestsVersion[floor][0]
		downV := gs.globalState.HallRequestsVersion[floor][1]

		fmt.Printf("Floor %d | Up: %v (v%d) | Down: %v (v%d)\n",
			floor, up, upV, down, downV)
	}

	// ---------------- Elevator States ----------------
	fmt.Println("\nElevator States:")

	for id, state := range gs.globalState.States {

		fmt.Printf("\nElevator %s\n", id)
		fmt.Printf("  Behavior:  %s\n", state.Behavior)
		fmt.Printf("  Floor:     %d\n", state.Floor)
		fmt.Printf("  Direction: %s\n", state.Direction)

		fmt.Printf("  CabRequests: ")
		for floor, active := range state.CabRequests {
			if active {
				fmt.Printf("[%d] ", floor)
			}
		}
		fmt.Println()
	}

	fmt.Println("==================================")
}