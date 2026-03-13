package state

import (
	"fmt"
	"heislab/elevator/elevio"
	"heislab/hallRequestAssigner"
	"heislab/management"
	"sync"
)

type GlobalStateData struct {
	HallRequests        [][2]bool                                        // [floor][0=up,1=down]
	HallRequestsVersion [][2]int                                         // incremented by one when matching hallRequest is updated
	States              map[string]hallRequestAssigner.ElevatorStateJSON // elevatorID -> state
	LocalID             string
}

type GlobalState struct {
	mu          sync.Mutex
	data        GlobalStateData
}

func InitGlobalState(elev *management.Elevator, elevID string) *GlobalState {
	gs := &GlobalState{}

	gs.data.HallRequests = make([][2]bool, management.NumFloors)
	gs.data.HallRequestsVersion = make([][2]int, management.NumFloors)
	gs.data.States = make(map[string]hallRequestAssigner.ElevatorStateJSON)
	gs.data.LocalID = elevID
	gs.data.States[elevID] = ConvertElevatorToJSON(elev) 
	return gs
}

// -------------------- State Conversion --------------------

func ConvertElevatorToJSON(elev *management.Elevator) hallRequestAssigner.ElevatorStateJSON {
	cabRequests := make([]bool, management.NumFloors)
	for floor := 0; floor < management.NumFloors; floor++ {
		cabRequests[floor] = elev.Orders[floor][management.CabButton].OrderPlaced
	}

	return hallRequestAssigner.ElevatorStateJSON{
		Behavior:    convertState(elev.State),
		Floor:       elev.LastFloor,
		Direction:   convertDirection(elev.MoveDir),
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

func (gs *GlobalState) UpdateGlobalState(elev *management.Elevator) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.data.States[elev.ID] = ConvertElevatorToJSON(elev)
}

func (gs *GlobalState) AddHallRequest(order management.Order) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.data.HallRequests[order.Floor][order.ButtonType] = true
}

func (gs *GlobalState) IncrementHallRequestVersion(floor int, btn elevio.ButtonType) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.data.HallRequestsVersion[floor][btn]++
}

func (gs *GlobalState) SetElevatorToOffline(deadID string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	elevState, exists := gs.data.States[deadID]
	if !exists {
		return
	}
	elevState.Behavior = "offline"
	gs.data.States[deadID] = elevState
}

// SetElevatorState setter en spesifikk elevator state i globalState
func (gs *GlobalState) SetElevatorGlobalState(elevID string, state hallRequestAssigner.ElevatorStateJSON) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.data.States[elevID] = state
}

// GetElevatorState henter state for en spesifikk heis
func (gs *GlobalState) GetElevatorState(elevID string) (hallRequestAssigner.ElevatorStateJSON, bool) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	state, ok := gs.data.States[elevID]
	return state, ok
}


func (gs *GlobalState) GetCopy() GlobalStateData {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	copyState := gs.data

	// deep copy slices
	copyState.HallRequests = append([][2]bool(nil), gs.data.HallRequests...)
	copyState.HallRequestsVersion = append([][2]int(nil), gs.data.HallRequestsVersion...)

	newMap := make(map[string]hallRequestAssigner.ElevatorStateJSON)
	for k, v := range gs.data.States {
		newMap[k] = v
	}
	copyState.States = newMap

	return copyState
}

func (gs *GlobalState) GetID() string {
    gs.mu.Lock()
    defer gs.mu.Unlock()

    return gs.data.LocalID
}

func (gs *GlobalState) RemoveHallRequest(floor int, btn elevio.ButtonType) {
	gs.mu.Lock()
	gs.data.HallRequests[floor][btn] = false
	gs.data.HallRequestsVersion[floor][btn]++
	gs.mu.Unlock()
}

func (gs *GlobalState) PrintGlobalState() {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	fmt.Println("\n========== GLOBAL STATE ==========")
	fmt.Printf("LocalID: %s\n", gs.data.LocalID)

	// ---------------- Hall Requests ----------------
	fmt.Println("\nHall Requests:")
	for floor := 0; floor < len(gs.data.HallRequests); floor++ {
		up := gs.data.HallRequests[floor][0]
		down := gs.data.HallRequests[floor][1]
		upV := gs.data.HallRequestsVersion[floor][0]
		downV := gs.data.HallRequestsVersion[floor][1]

		fmt.Printf("Floor %d | Up: %v (v%d) | Down: %v (v%d)\n",
			floor, up, upV, down, downV)
	}

	// ---------------- Elevator States ----------------
	fmt.Println("\nElevator States:")

	for id, state := range gs.data.States {

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