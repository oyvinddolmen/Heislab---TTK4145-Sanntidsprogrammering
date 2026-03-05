package orderManagement

import (
	"fmt"
	"heislab/elevio"
	"heislab/management"
)

// ---------------------------------------------------------------------
// UpdateCurrentOrder: velger neste ordre for heisen basert på localElevator
// ---------------------------------------------------------------------
func UpdateCurrentOrder() {
	e := &management.Elev
	fmt.Println("Movedir før UpdatecurrentOrder ", e.MoveDir)

	// Hvis vi allerede har en aktiv ordre → behold den
	//if e.CurrentOrder.OrderPlaced {
	//	fmt.Println("returnerte ingenting")
	//	return
	//}

	// Start fra gjeldende etasje, eller siste etasje hvis mellom etasjer
	startFloor := e.Floor
	if startFloor < 0 {
		startFloor = e.LastFloor
		if startFloor < 0 {
			startFloor = 0
		}
	}

	found := false
	switch e.MoveDir {

	case management.DirUp:
		found = assignUp(e, startFloor)
		if !found {
			found = assignDown(e, startFloor)
		}

	case management.DirDown:
		found = assignDown(e, startFloor)
		if !found {
			found = assignUp(e, startFloor)
		}

	default: // Idle
		found = assignUp(e, startFloor)
		if !found {
			found = assignDown(e, startFloor)
		}
	}

	// Ingen ordre funnet → sett heis til idle
	if !found {
		management.Elev.State = management.ElevIdle
		fmt.Println("Ordre ikke found: ")
	}
}

// assignUp: finn første ordre oppover fra startFloor
// ---------------------------------------------------------------------
func assignUp(e *management.Elevator, startFloor int) bool {
	for floor := startFloor; floor < management.NumFloors; floor++ {
		for button := 0; button < management.NumButtons; button++ {
			order := &e.Orders[floor][button]
			fmt.Println("Ordresjekk:  True? ", order.OrderPlaced , "Floor: ", order.Floor)
			if order.OrderPlaced {
				Ordre := *order
				fmt.Println("Ordre satt: ", Ordre.Floor)
				e.CurrentOrder = *order
				return true
			}
		}
	}
	return false
}

// assignDown: finn første ordre nedover fra startFloor
// ---------------------------------------------------------------------
func assignDown(e *management.Elevator, startFloor int) bool {
	if startFloor >= management.NumFloors {
		startFloor = management.NumFloors - 1
	}

	for floor := startFloor; floor >= 0; floor-- {
		for button := 0; button < management.NumButtons; button++ {
			order := &e.Orders[floor][button]
			if order.OrderPlaced {
				Ordre := *order
				fmt.Println("Ordre satt: ", Ordre.Floor)
				e.CurrentOrder = *order
				return true
			}
		}
	}
	return false
}

// ---------------------------------------------------------------------
// call when elevator has reached CurrentOrder.Floor
// ---------------------------------------------------------------------
func CompleteCurrentOrder(gs *GlobalState) {
	elevator := &management.Elev
	floor := elevator.CurrentOrder.Floor
	button := elevator.CurrentOrder.ButtonType

	// --- CAB ORDER ---
	if button == elevio.CabButton {

		// Oppdater lokal heis
		//elevator.Orders[floor][button].Finished = true
		elevator.Orders[floor][button].OrderPlaced = false
		//elevator.CurrentOrder.Finished = true
		elevator.CurrentOrder.OrderPlaced = false

		gs.UpdateLocalGlobalState()

	} else {
		// --- HALL ORDER ---
		//elevator.Orders[floor][button].Finished = true
		//elevator.CurrentOrder.Finished = true
		elevator.CurrentOrder.OrderPlaced = false

		gs.mu.Lock()
		gs.globalState.HallRequests[floor][button] = false
		gs.mu.Unlock()

		gs.IncrementHallRequestVersion(elevator.CurrentOrder)
	}
	UpdateCurrentOrder()
}