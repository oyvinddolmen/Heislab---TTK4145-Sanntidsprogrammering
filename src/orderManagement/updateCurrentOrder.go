package orderManagement

import (
	"heislab/elevio"
	"heislab/management"
<<<<<<< HEAD
	"fmt"
=======
>>>>>>> Oyvind
)

// ---------------------------------------------------------------------
// UpdateCurrentOrder: velger neste ordre for heisen basert på GlobalState
// ---------------------------------------------------------------------
func UpdateCurrentOrder() {
	e := &management.Elev

	// Hvis vi allerede har en aktiv ordre som ikke er ferdig: gjør ingenting
	if e.CurrentOrder.OrderPlaced && !e.CurrentOrder.Finished {
		return
	}

	// Start fra aktuell etasje, eller 0 hvis -1
	startFloor := e.Floor
	if startFloor < 0 {
		startFloor = 0
	}

	// Velg retning basert på tidligere retning
	switch e.MoveDir {
	case management.Dir_Up:
		if assignUp(startFloor) {
			return
		}
		if assignDown(startFloor) {
			return
		}
	case management.Dir_Down:
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
	e.State = management.IDLE
	e.MoveDir = management.Dir_Idle
}

// ---------------------------------------------------------------------
// assignUp: finn første ordre oppover fra startFloor
// ---------------------------------------------------------------------
func assignUp(startFloor int) bool {
	e := &management.Elev

	for f := startFloor; f < management.NumFloors; f++ {
		for b := 0; b < management.NumButtons; b++ {
			order := &e.Orders[f][b]
			if order.OrderPlaced && !order.Finished {
				e.CurrentOrder = *order
				e.MoveDir = management.Dir_Up
				e.State = management.MOVING
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
	e := &management.Elev

	if startFloor >= management.NumFloors {
		startFloor = management.NumFloors - 1
	}

	for f := startFloor; f >= 0; f-- {
		for b := 0; b < management.NumButtons; b++ {
			order := &e.Orders[f][b]
			if order.OrderPlaced && !order.Finished {
				e.CurrentOrder = *order
				e.MoveDir = management.Dir_Down
				e.State = management.MOVING
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

	e := &management.Elev
	f := e.CurrentOrder.Floor
	b := e.CurrentOrder.ButtonType
<<<<<<< HEAD
=======
	localID := e.IP
>>>>>>> Oyvind

	// --- CAB ORDER ---
	if b == elevio.BT_Cab {

		// Oppdater lokal heis
		e.Orders[f][b].Finished = true
		e.CurrentOrder.Finished = true
		e.CurrentOrder.OrderPlaced = false

		UpdateLocalGlobalState()

	} else {

		// --- HALL ORDER ---
		e.Orders[f][b].Finished = true
		e.CurrentOrder.Finished = true
		e.CurrentOrder.OrderPlaced = false

		GlobalStateMutex.Lock()
		GlobalState.HallRequests[f][b] = false
		GlobalStateMutex.Unlock()

		IncremtHallRequestVersion(e.CurrentOrder)
		fmt.Print("+ på Version, ordre FULLFØRT")
	}

	UpdateCurrentOrder()
}
