package state

import (
	"fmt"
	"heislab/elevator/elevIO"
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
	mutex       sync.Mutex
	data        GlobalStateData
}

func InitGlobalState(elev *management.Elevator, elevID string) *GlobalState {
	globalState := &GlobalState{}

	globalState.data.HallRequests = make([][2]bool, management.NumFloors)
	globalState.data.HallRequestsVersion = make([][2]int, management.NumFloors)
	globalState.data.States = make(map[string]hallRequestAssigner.ElevatorStateJSON)
	globalState.data.LocalID = elevID
	globalState.data.States[elevID] = ConvertElevatorToJSON(elev) 
	return globalState
}

// -------------------------------------------------------------------------------------------
// State conversion
// -------------------------------------------------------------------------------------------

func ConvertElevatorToJSON(elev *management.Elevator) hallRequestAssigner.ElevatorStateJSON {
	cabRequests := make([]bool, management.NumFloors)
	for floor := 0; floor < management.NumFloors; floor++ {
		cabRequests[floor] = elev.Orders[floor][management.CabButton].OrderPlaced
	}

	return hallRequestAssigner.ElevatorStateJSON{
		Behavior:      convertState(elev.State),
		Floor:         elev.LastFloor,
		Direction:     convertDirection(elev.MoveDir),
		CabRequests:   cabRequests,
		CanTakeOrders: elev.CanTakeOrders,
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

// -------------------------------------------------------------------------------------------
// Public methods
// -------------------------------------------------------------------------------------------

func (globalState *GlobalState) UpdateGlobalState(elev *management.Elevator) {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()
	globalState.data.States[elev.ID] = ConvertElevatorToJSON(elev)
}

func (globalState *GlobalState) AddHallRequest(order management.Order) {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()
	globalState.data.HallRequests[order.Floor][order.ButtonType] = true
}

func (globalState *GlobalState) IncrementHallRequestVersion(floor int, btn elevIO.ButtonType) {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()
	globalState.data.HallRequestsVersion[floor][btn]++
}

func (globalState *GlobalState) SetElevatorToOffline(deadID string) {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()
	elevState, exists := globalState.data.States[deadID]
	if !exists {
		return
	}
	elevState.Behavior = "offline"
	globalState.data.States[deadID] = elevState
}

// SetElevatorState setter en spesifikk elevator state i globalState
func (globalState *GlobalState) SetElevatorGlobalState(elevID string, state hallRequestAssigner.ElevatorStateJSON) {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()
	globalState.data.States[elevID] = state
}

// GetElevatorState henter state for en spesifikk heis
func (globalState *GlobalState) GetElevatorState(elevID string) (hallRequestAssigner.ElevatorStateJSON, bool) {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()
	state, ok := globalState.data.States[elevID]
	return state, ok
}


func (globalState *GlobalState) GetCopy() GlobalStateData {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()

	copyState := globalState.data

	// deep copy slices
	copyState.HallRequests = append([][2]bool(nil), globalState.data.HallRequests...)
	copyState.HallRequestsVersion = append([][2]int(nil), globalState.data.HallRequestsVersion...)

	newMap := make(map[string]hallRequestAssigner.ElevatorStateJSON)
	for k, v := range globalState.data.States {
		newMap[k] = v
	}
	copyState.States = newMap

	return copyState
}

func (globalState *GlobalState) GetLocalID() string {
    globalState.mutex.Lock()
    defer globalState.mutex.Unlock()

    return globalState.data.LocalID
}

func (globalState *GlobalState) RemoveHallRequest(floor int, button elevIO.ButtonType) {
	globalState.mutex.Lock()
	globalState.data.HallRequests[floor][button] = false
	globalState.data.HallRequestsVersion[floor][button]++
	globalState.mutex.Unlock()
}

func (globalState *GlobalState) PrintGlobalState() {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()

	fmt.Println("\n========== GLOBAL STATE ==========")
	fmt.Printf("LocalID: %s\n", globalState.data.LocalID)

	// ---------------- Hall Requests ----------------
	fmt.Println("\nHall Requests:")
	for floor := 0; floor < len(globalState.data.HallRequests); floor++ {
		up := globalState.data.HallRequests[floor][0]
		down := globalState.data.HallRequests[floor][1]
		upV := globalState.data.HallRequestsVersion[floor][0]
		downV := globalState.data.HallRequestsVersion[floor][1]

		fmt.Printf("Floor %d | Up: %v (v%d) | Down: %v (v%d)\n",
			floor, up, upV, down, downV)
	}

	// ---------------- Elevator States ----------------
	fmt.Println("\nElevator States:")

	for id, state := range globalState.data.States {

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