package orderManagement

import (
	"fmt"
	"heislab/elevio"
	"heislab/management"
)

// ---------------------------------------------------------------------
// UpdateCurrentOrder: velger neste ordre for heisen basert på GlobalState
// ---------------------------------------------------------------------
func UpdateCurrentOrder() {
	elevator := &management.Elev

	// Hvis vi allerede har en aktiv ordre som ikke er ferdig: gjør ingenting
	if elevator.CurrentOrder.OrderPlaced && !elevator.CurrentOrder.Finished {
		return
	}

	// Start fra aktuell etasje, eller 0 hvis -1
	startFloor := elevator.Floor
	if startFloor < 0 {
		startFloor = 0
	}

	// Velg retning basert på tidligere retning
	switch elevator.MoveDir {
	case management.DirUp:
		if assignUp(startFloor) {
			return
		}
		if assignDown(startFloor) {
			return
		}
	case management.DirDown:
		if assignDown(startFloor) {
			return
		}
		if assignUp(startFloor) {
			return
		}
	default: // Dir_Idle eller stopper
		if assignUp(startFloor) {
			return
		}
		if assignDown(startFloor) {
			return
		}
	}

	// Ingen ordre funnet → gå IDLE
	elevator.State = management.ElevIdle
	elevator.MoveDir = management.DirIdle
}

// ---------------------------------------------------------------------
// assignUp: finn første ordre oppover fra startFloor
// ---------------------------------------------------------------------
func assignUp(startFloor int) bool {
	elevator := &management.Elev

	for floor := startFloor; floor < management.NumFloors; floor++ {
		for button := 0; button < management.NumButtons; button++ {
			order := &elevator.Orders[floor][button]
			if order.OrderPlaced && !order.Finished {
				elevator.CurrentOrder = *order
				elevator.MoveDir = management.DirUp
				elevator.State = management.ElevMoving
				return true
			}
		}
	}
	return false
}

// ---------------------------------------------------------------------
// assignDown: finn første ordre nedover fra startFloor
// ---------------------------------------------------------------------
func assignDown(startFloor int) bool {
	elevator := &management.Elev

	if startFloor >= management.NumFloors {
		startFloor = management.NumFloors - 1
	}

	for floor := startFloor; floor >= 0; floor-- {
		for button := 0; button < management.NumButtons; button++ {
			order := &elevator.Orders[floor][button]
			if order.OrderPlaced && !order.Finished {
				elevator.CurrentOrder = *order
				elevator.MoveDir = management.DirDown
				elevator.State = management.ElevMoving
				return true
			}
		}
	}
	return false
}

// ---------------------------------------------------------------------
// call when elevator has reached CurrentOrder.Floor
// ---------------------------------------------------------------------
func CompleteCurrentOrder() {

	elevator := &management.Elev
	floor := elevator.CurrentOrder.Floor
	button := elevator.CurrentOrder.ButtonType

	// --- CAB ORDER ---
	if button == elevio.BT_Cab {

		// Oppdater lokal heis
		elevator.Orders[floor][button].Finished = true
		elevator.Orders[floor][button].OrderPlaced = false
		elevator.CurrentOrder.Finished = true
		elevator.CurrentOrder.OrderPlaced = false

		UpdateLocalGlobalState()

	} else {

		// --- HALL ORDER ---
		elevator.Orders[floor][button].Finished = true
		elevator.CurrentOrder.Finished = true
		elevator.CurrentOrder.OrderPlaced = false

		GlobalStateMutex.Lock()
		GlobalState.HallRequests[floor][button] = false
		GlobalStateMutex.Unlock()

		IncremtHallRequestVersion(elevator.CurrentOrder)
		fmt.Print("+ på Version, ordre FULLFØRT")
	}

	UpdateCurrentOrder()
}
