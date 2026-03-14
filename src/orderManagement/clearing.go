package orderManagement

import (
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/state"
)

// TODO: COMMENT WHAT DOES THIS FUNCTION DO??
func ClearOrdersAndTurnOffLights(elev *management.Elevator, globalState *state.GlobalState) bool {
	currentFloor := elev.Floor
	if elev.CurrentOrder.Floor == currentFloor {
		elev.CurrentOrder.OrderPlaced = false
		elev.LastOrder = elev.CurrentOrder
	}

	if elev.Orders[currentFloor][elevIO.CabButton].OrderPlaced {
		RemoveCabOrder(globalState, elev, currentFloor)
	}

	switch elev.MoveDir {

	case management.DirUp:

		if elev.Orders[currentFloor][elevIO.HallUpButton].OrderPlaced {
			RemoveHallUp(globalState, elev, currentFloor)
		}

	case management.DirDown:

		if elev.Orders[currentFloor][elevIO.HallDownButton].OrderPlaced {
			RemoveHallDown(globalState, elev, currentFloor)
		}

	default:

		if elev.Orders[currentFloor][elevIO.HallUpButton].OrderPlaced &&
			elev.LastOrder.ButtonType == elevIO.HallDownButton &&
			CabOrderAbove(elev, currentFloor) {
			RemoveHallUp(globalState, elev, currentFloor)
			return true

		} else if elev.Orders[currentFloor][elevIO.HallDownButton].OrderPlaced &&
			elev.LastOrder.ButtonType == elevIO.HallUpButton &&
			CabOrderBelow(elev, currentFloor) {
			RemoveHallDown(globalState, elev, currentFloor)
			return true
		}
	}

	return false
}

func CabOrderAbove(elev *management.Elevator, currentFloor int) bool {
	for floor := currentFloor + 1; floor < management.NumFloors; floor++ {
		if elev.Orders[floor][elevIO.CabButton].OrderPlaced {
			return true
		}
	}
	return false
}

func CabOrderBelow(elev *management.Elevator, currentFloor int) bool {
	for floor := currentFloor - 1; floor >= 0; floor-- {
		if elev.Orders[floor][elevIO.CabButton].OrderPlaced {
			return true
		}
	}
	return false
}

func OrdersAbove(elev *management.Elevator, floorUnderInspection int) bool {
	if floorUnderInspection == management.NumFloors-1 {
		return false
	}

	for floor := floorUnderInspection + 1; floor < management.NumFloors; floor++ {
		for button := 0; button < management.NumButtons; button++ {
			if elev.Orders[floor][button].OrderPlaced {
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

	for floor := floorUnderInspection - 1; floor >= 0; floor-- {
		for button := 0; button < management.NumButtons; button++ {
			if elev.Orders[floor][button].OrderPlaced {
				return true
			}
		}
	}
	return false
}

func RemoveCabOrder(globalState *state.GlobalState, elev *management.Elevator, floor int) {
	elev.Orders[floor][elevIO.CabButton].OrderPlaced = false
	globalState.UpdateGlobalState(elev)
	elevIO.SetButtonLamp(elevIO.CabButton, floor, false)
}

func RemoveHallDown(globalState *state.GlobalState, elev *management.Elevator, floor int) {
	elev.Orders[floor][elevIO.HallDownButton].OrderPlaced = false
	globalState.RemoveHallRequest(floor, elevIO.HallDownButton)
	elevIO.SetButtonLamp(elevIO.HallDownButton, floor, false)
}

func RemoveHallUp(globalState *state.GlobalState, elev *management.Elevator, floor int) {
	elev.Orders[floor][elevIO.HallUpButton].OrderPlaced = false
	globalState.RemoveHallRequest(floor, elevIO.HallUpButton)
	elevIO.SetButtonLamp(elevIO.HallUpButton, floor, false)
}

// Clears hall requests from the shared state for the floor the elevator is currently at
func ServeHallRequestsAtCurrentFloor(elev *management.Elevator, globalState *state.GlobalState) {
	currentFloor := elev.Floor
	if currentFloor != -1 {
		hallRequests := globalState.GetCopy().HallRequests

		if hallRequests[currentFloor][elevIO.HallUpButton] {
			RemoveHallUp(globalState, elev, currentFloor)
		}
		if hallRequests[currentFloor][elevIO.HallDownButton] {
			RemoveHallDown(globalState, elev, currentFloor)
		}

	}
}
