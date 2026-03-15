package state

import (
	"fmt"
	"heislab/elevator/elevIO"
	"heislab/management"
	"sync"
)

type GlobalStateData struct {
	HallOrders          [][management.NumHallButtonTypes]bool                                        // [floor][0=up,1=down]
	HallOrderVersion    [][management.NumHallButtonTypes]int                                         // incremented by one when matching hallOrder is updated
	States              map[string]ElevatorStateJSON // elevatorID -> state
	LocalID             string
}

type ElevatorStateJSON struct {
	Behavior      string `json:"behaviour"` // idle, moving, doorOpen, offline
	Floor         int    `json:"floor"`
	Direction     string `json:"direction"`
	CabOrders     []bool `json:"cabOrders"`
	CanTakeOrders bool   `json:"canTakeOrders"`
}

type GlobalState struct {
	mutex       sync.Mutex
	data        GlobalStateData
}

func InitGlobalState(elev *management.Elevator, elevID string) *GlobalState {
	globalState := &GlobalState{}

	globalState.data.HallOrders = make([][management.NumHallButtonTypes]bool, management.NumFloors)
	globalState.data.HallOrderVersion = make([][management.NumHallButtonTypes]int, management.NumFloors)
	globalState.data.States = make(map[string]ElevatorStateJSON)
	globalState.data.LocalID = elevID
	globalState.data.States[elevID] = ConvertElevatorToJSON(elev) 
	return globalState
}

// -------------------------------------------------------------------------------------------
// State conversion
// -------------------------------------------------------------------------------------------

func ConvertElevatorToJSON(elev *management.Elevator) ElevatorStateJSON {
	cabOrders := make([]bool, management.NumFloors)
	for floor := 0; floor < management.NumFloors; floor++ {
		cabOrders[floor] = elev.GetOrderActiveStatus(floor, management.CabButton)
	}

	return ElevatorStateJSON{
		Behavior:      convertState(elev.GetState()),
		Floor:         elev.GetFloor(),
		Direction:     convertDirection(elev.GetMoveDir()),
		CabOrders:     cabOrders,
		CanTakeOrders: elev.GetCanTakeOrders(),
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
	globalState.data.States[elev.GetID()] = ConvertElevatorToJSON(elev)
}

func (globalState *GlobalState) AddHallOrder(order management.Order) {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()
	globalState.data.HallOrders[order.GetFloor()][order.GetButtonType()] = true
}

func (globalState *GlobalState) IncrementHallOrderVersion(floor int, button elevIO.ButtonType) {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()
	globalState.data.HallOrderVersion[floor][button]++
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
func (globalState *GlobalState) SetElevatorGlobalState(elevID string, state ElevatorStateJSON) {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()
	globalState.data.States[elevID] = state
}

// GetElevatorState henter state for en spesifikk heis
func (globalState *GlobalState) GetElevatorState(elevID string) (ElevatorStateJSON, bool) {
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
	copyState.HallOrders = append([][management.NumHallButtonTypes]bool(nil), globalState.data.HallOrders...)
	copyState.HallOrderVersion = append([][management.NumHallButtonTypes]int(nil), globalState.data.HallOrderVersion...)

	newMap := make(map[string]ElevatorStateJSON)
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

func (globalState *GlobalState) RemoveHallOrder(floor int, button elevIO.ButtonType) {
	globalState.mutex.Lock()
	globalState.data.HallOrders[floor][button] = false
	globalState.data.HallOrderVersion[floor][button]++
	globalState.mutex.Unlock()
}

func (globalState *GlobalState) PrintGlobalState() {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()

	fmt.Println("\n========== GLOBAL STATE ==========")
	fmt.Printf("LocalID: %s\n", globalState.data.LocalID)

	// ---------------- Hall Orders ----------------
	fmt.Println("\nHall Orders:")
	for floor := 0; floor < len(globalState.data.HallOrders); floor++ {
		hallUp := globalState.data.HallOrders[floor][0]
		hallDown := globalState.data.HallOrders[floor][1]
		hallUpVersion := globalState.data.HallOrderVersion[floor][0]
		hallDownVersion := globalState.data.HallOrderVersion[floor][1]

		fmt.Printf("Floor %d | Up: %v (v%d) | Down: %v (v%d)\n",
			floor, hallUp, hallUpVersion, hallDown, hallDownVersion)
	}

	// ---------------- Elevator States ----------------
	fmt.Println("\nElevator States:")

	for elevID, state := range globalState.data.States {

		fmt.Printf("\nElevator %s\n", elevID)
		fmt.Printf("  Behavior:  %s\n", state.Behavior)
		fmt.Printf("  Floor:     %d\n", state.Floor)
		fmt.Printf("  Direction: %s\n", state.Direction)

		fmt.Printf("  CabOrders: ")
		for floor, active := range state.CabOrders {
			if active {
				fmt.Printf("[%d] ", floor)
			}
		}
		fmt.Println()
	}

	fmt.Println("==================================")
}