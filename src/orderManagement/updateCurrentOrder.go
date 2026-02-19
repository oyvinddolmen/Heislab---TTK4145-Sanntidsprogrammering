package orderManagement

import "heislab/management"

// ---------------------------------------------------------------------
// Update the CurrentOrder for the global elevator Elev
// ---------------------------------------------------------------------
func UpdateCurrentOrder() {

	e := &management.Elev // bruk global heis

	// Hvis vi allerede har en ordre som ikke er fullført, behold den
	if e.CurrentOrder.OrderPlaced && !e.CurrentOrder.Finished {
		return
	}

	switch e.MoveDir {

	case management.Dir_Up:
		if assignUp() {
			return
		}
		if assignDown() {
			return
		}

	case management.Dir_Down:
		if assignDown() {
			return
		}
		if assignUp() {
			return
		}

	default: // Dir_Idle eller stopper
		if assignUp() {
			return
		}
		if assignDown() {
			return
		}
	}

	// Ingen ordre funnet
	e.State = management.IDLE
	e.MoveDir = management.Dir_Idle
}

// ---------------------------------------------------------------------
// assignUp: finn første ordre oppover fra nåværende etasje
// ---------------------------------------------------------------------
func assignUp() bool {

	e := &management.Elev

	for f := e.Floor; f < management.NumFloors; f++ {
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
// assignDown: finn første ordre nedover fra nåværende etasje
// ---------------------------------------------------------------------
func assignDown() bool {

	e := &management.Elev

	for f := e.Floor; f >= 0; f-- {
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
// CompleteCurrentOrder: kall når heisen har nådd CurrentOrder.Floor
// ---------------------------------------------------------------------
func CompleteCurrentOrder() {

	e := &management.Elev

	f := e.CurrentOrder.Floor
	b := e.CurrentOrder.ButtonType

	e.Orders[f][b].Finished = true
	e.CurrentOrder.Finished = true
	e.CurrentOrder.OrderPlaced = false

	UpdateCurrentOrder()
}
