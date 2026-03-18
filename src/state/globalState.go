package state

import (
	"heislab/elevator/elevIO"
	"heislab/management"
	"sync"
)

type GlobalStateData struct {
	HallOrders       [][management.NumHallButtonTypes]bool // [floor][0=up,1=down]
	HallOrderVersion [][management.NumHallButtonTypes]int  // incremented by one when matching hallOrder is updated
	States           map[string]ElevatorStateJSON          // elevatorID -> state
	LocalID          string
}

type GlobalState struct {
	mutex sync.Mutex
	data  GlobalStateData
}

// Creates and initializes a GlobalState for the local elevator.
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

func (globalState *GlobalState) SetElevatorGlobalState(elevID string, elevState ElevatorStateJSON) {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()
	globalState.data.States[elevID] = elevState
}

func (globalState *GlobalState) GetElevatorState(elevID string) (ElevatorStateJSON, bool) {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()
	elevState, exists := globalState.data.States[elevID]
	return elevState, exists
}

func (globalState *GlobalState) GetCopy() GlobalStateData {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()

	globalStateCopy := globalState.data

	// deep copy slices
	globalStateCopy.HallOrders = append([][management.NumHallButtonTypes]bool(nil), globalState.data.HallOrders...)
	globalStateCopy.HallOrderVersion = append([][management.NumHallButtonTypes]int(nil), globalState.data.HallOrderVersion...)

	elevStatesCopy := make(map[string]ElevatorStateJSON)
	for elevID, elevState := range globalState.data.States {
		elevStatesCopy[elevID] = elevState
	}
	globalStateCopy.States = elevStatesCopy

	return globalStateCopy
}

func (globalState *GlobalState) GetLocalID() string {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()
	return globalState.data.LocalID
}

func (globalState *GlobalState) RemoveHallOrder(floor int, button elevIO.ButtonType) {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()
	globalState.data.HallOrders[floor][button] = false
	globalState.data.HallOrderVersion[floor][button]++
}

// -------------------------------------------------------------------------------------------
// Failure Detection Helpers
// -------------------------------------------------------------------------------------------

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

func (globalState *GlobalState) IsOffline() bool {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()

	localID := globalState.data.LocalID
	state, exists := globalState.data.States[localID]
	if !exists {
		return false
	}
	return state.Behavior == "offline"
}

func (globalState *GlobalState) SetSelfToOnline(elev *management.Elevator) {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()

	localID := globalState.data.LocalID
	state, exists := globalState.data.States[localID]
	if !exists {
		return
	}
	state.Behavior = convertState(elev.GetState())
	globalState.data.States[localID] = state
}

func (globalState *GlobalState) AllOtherElevatorsOffline() bool {
	globalState.mutex.Lock()
	defer globalState.mutex.Unlock()

	localID := globalState.data.LocalID

	for id, s := range globalState.data.States {
		if id == localID {
			continue
		}
		if s.Behavior != "offline" {
			return false
		}
	}
	return true
}
