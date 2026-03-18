package orderManagement

import (
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/state"
)

// TODO: Burde dette vært to forskjellige funksjoner?
// Clears orders at current floor and returns true if there is a HallOrder conflict ()
func ClearOrdersAtCurrentFloor(elev *management.Elevator, globalState *state.GlobalState){
	currentFloor := elev.GetFloor()
	if elev.GetCurrentOrderFloor() == currentFloor {
		elev.SetCurrentOrderActiveStatus(false)
		elev.SetLastOrder(elev.GetCurrentOrder())
	}

	if elev.GetOrderActiveStatus(currentFloor, int(elevIO.CabButton)) {
		RemoveCabOrder(globalState, elev, currentFloor)
	}

	switch elev.GetMoveDir() {

	case management.DirUp:

		if elev.GetOrderActiveStatus(currentFloor, int(elevIO.HallUpButton)) {
			RemoveHallUp(globalState, elev, currentFloor)
		}

	case management.DirDown:

		if elev.GetOrderActiveStatus(currentFloor, int(elevIO.HallDownButton)) {
			RemoveHallDown(globalState, elev, currentFloor)
		}

	default:
		// If active HallUp order at current floor and active Cab order above, and last order was HallDown.
		if elev.GetOrderActiveStatus(currentFloor, int(elevIO.HallUpButton)) &&
			elev.GetLastOrder().IsHallDownOrder() && 
			CabOrderAbove(elev, currentFloor) {
			RemoveHallUp(globalState, elev, currentFloor)
		
		// Else if active HallDown order at current floor and active Cab order below, and last order was HallUp.
		} else if elev.GetOrderActiveStatus(currentFloor, int(elevIO.HallDownButton)) &&
			elev.GetLastOrder().IsHallUpOrder() &&
			CabOrderBelow(elev, currentFloor) {
			RemoveHallDown(globalState, elev, currentFloor)
		}
	}
}

func CabOrderAbove(elev *management.Elevator, currentFloor int) bool {
	for floor := currentFloor + 1; floor < management.NumFloors; floor++ {
		if elev.GetOrderActiveStatus(floor, int(elevIO.CabButton)) {
			return true
		}
	}
	return false
}

func CabOrderBelow(elev *management.Elevator, currentFloor int) bool {
	for floor := currentFloor - 1; floor >= 0; floor-- {
		if elev.GetOrderActiveStatus(floor, int(elevIO.CabButton)) {
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
			if elev.GetOrderActiveStatus(floor, button) {
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
			if elev.GetOrderActiveStatus(floor, button) {
				return true
			}
		}
	}
	return false
}

func RemoveCabOrder(globalState *state.GlobalState, elev *management.Elevator, floor int) {
	elev.SetOrderActiveStatus(floor, int(elevIO.CabButton), false)
	globalState.UpdateGlobalState(elev)
	elevIO.SetButtonLamp(elevIO.CabButton, floor, false)
}

func RemoveHallDown(globalState *state.GlobalState, elev *management.Elevator, floor int) {
	elev.SetOrderActiveStatus(floor, int(elevIO.HallDownButton), false)
	globalState.RemoveHallOrder(floor, elevIO.HallDownButton)
	elevIO.SetButtonLamp(elevIO.HallDownButton, floor, false)
}

func RemoveHallUp(globalState *state.GlobalState, elev *management.Elevator, floor int) {
	elev.SetOrderActiveStatus(floor, int(elevIO.HallUpButton), false)
	globalState.RemoveHallOrder(floor, elevIO.HallUpButton)
	elevIO.SetButtonLamp(elevIO.HallUpButton, floor, false)
}

// Clears hall orders from the shared state for the floor the elevator is currently at
func ServeHallOrdersAtCurrentFloor(elev *management.Elevator, globalState *state.GlobalState) {
	currentFloor := elev.GetFloor()
	if currentFloor != -1 {
		hallOrders := globalState.GetCopy().HallOrders

		if hallOrders[currentFloor][elevIO.HallUpButton] {
			RemoveHallUp(globalState, elev, currentFloor)
		}
		if hallOrders[currentFloor][elevIO.HallDownButton] {
			RemoveHallDown(globalState, elev, currentFloor)
		}

	}
}
