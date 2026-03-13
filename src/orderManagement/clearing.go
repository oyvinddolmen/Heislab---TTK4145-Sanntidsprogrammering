
package orderManagement

import (
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/state"
)

func ClearOrdersAndTurnOfLights(elev *management.Elevator, gs *state.GlobalState) bool {
	floor := elev.Floor
	if elev.CurrentOrder.Floor == floor {
		elev.CurrentOrder.OrderPlaced = false
		elev.LastOrder = elev.CurrentOrder
	}
	// CAB orders are always removed
	if elev.Orders[floor][elevIO.CabButton].OrderPlaced {
		RemoveCabOrder(gs, elev, floor)
	}

	switch elev.MoveDir {

	case management.DirUp:

		if elev.Orders[floor][elevIO.HallUpButton].OrderPlaced {
			RemoveHallUp(gs, elev, floor)
		}

	case management.DirDown:

		if elev.Orders[floor][elevIO.HallDownButton].OrderPlaced {
			RemoveHallDown(gs, elev, floor)
		}

	default:

	if elev.Orders[floor][elevIO.HallUpButton].OrderPlaced &&
		elev.LastOrder.ButtonType == elevIO.HallDownButton &&
		CabOrderAbove(elev, floor) {
		RemoveHallUp(gs, elev, floor)
		return true

	} else if elev.Orders[floor][elevIO.HallDownButton].OrderPlaced &&
				elev.LastOrder.ButtonType == elevIO.HallUpButton &&
				CabOrderBelow(elev, floor) {
					RemoveHallDown(gs, elev, floor)
					return true
				}
	}
	
	return false
}

func CabOrderAbove(elev *management.Elevator, floor int) bool {
	for f := floor + 1; f < management.NumFloors; f++ {
		if elev.Orders[f][elevIO.CabButton].OrderPlaced {
			return true
		}
	}
	return false
}

func CabOrderBelow(elev *management.Elevator, floor int) bool {
	for f := floor - 1; f >= 0; f-- {
		if elev.Orders[f][elevIO.CabButton].OrderPlaced {
			return true
		}
	}
	return false
}

func OrdersAbove(elev *management.Elevator, floorUnderInspection int) bool {
	if floorUnderInspection == management.NumFloors-1 {
		return false
	}

	for f := floorUnderInspection + 1; f < management.NumFloors; f++ {
		for b := 0; b < management.NumButtons; b++ {
			if elev.Orders[f][b].OrderPlaced {
				return true
			}
		}
	}

	return false
}

func OrdersBelow(elev *management.Elevator, floorUnderInspection int) bool {
	if floorUnderInspection == 0 {
		return false
	}

	for f := floorUnderInspection - 1; f >= 0; f-- {
		for b := 0; b < management.NumButtons; b++ {
			if elev.Orders[f][b].OrderPlaced {
				return true
			}
		}
	}

	return false
}

func RemoveCabOrder(gs *state.GlobalState, elev *management.Elevator, floor int) {
	elev.Orders[floor][elevIO.CabButton].OrderPlaced = false
	gs.UpdateGlobalState(elev)
	elevIO.SetButtonLamp(elevIO.CabButton, floor, false)
}

func RemoveHallDown(gs *state.GlobalState, elev *management.Elevator, floor int) {

	elev.Orders[floor][elevIO.HallDownButton].OrderPlaced = false

	gs.RemoveHallRequest(floor, elevIO.HallDownButton)

	elevIO.SetButtonLamp(elevIO.HallDownButton, floor, false)
}

func RemoveHallUp(gs *state.GlobalState, elev *management.Elevator, floor int) {

	elev.Orders[floor][elevIO.HallUpButton].OrderPlaced = false

	gs.RemoveHallRequest(floor, elevIO.HallUpButton)

	elevIO.SetButtonLamp(elevIO.HallUpButton, floor, false)
}

// Clears hall requests from the shared state for the floor the elevator is currently at
func ServeHallRequestsAtCurrentFloor(elev *management.Elevator, gs *state.GlobalState) {
	floor := elev.Floor
	if floor != -1 {
		hallRequests := gs.GetCopy().HallRequests

		if hallRequests[floor][elevIO.HallUpButton] {
			RemoveHallUp(gs, elev, floor)
		}
		if hallRequests[floor][elevIO.HallDownButton] {
			RemoveHallDown(gs, elev, floor)
		}

	}
}
