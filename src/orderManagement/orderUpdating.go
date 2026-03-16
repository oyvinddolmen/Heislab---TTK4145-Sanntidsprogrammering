package orderManagement

import (
	"fmt"
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/state"
)

func UpdateCurrentOrder(elev *management.Elevator, globalState *state.GlobalState) {
	elev.PrintOrdersDebug()
	floor := elev.GetFloor()
	if elev.GetFloor() == -1 {
		return
	}

	switch elev.GetMoveDir() {
	case management.DirUp:
		if assignUp(globalState, elev, floor) {
			return
		}

		if OrdersBelow(elev, elev.GetFloor()) {
			assignDown(globalState, elev, floor)
			elev.SetMoveDir(management.DirDown)
			return
		}

	case management.DirDown:
		if assignDown(globalState, elev, floor) {
			return
		}

		if OrdersAbove(elev, elev.GetFloor()) {
			assignUp(globalState, elev, floor)
			elev.SetMoveDir(management.DirUp)
			return
		}

	default: // Idle
		if assignUp(globalState, elev, floor) {
			elev.SetMoveDir(management.DirUp)
			return
		}
		if assignDown(globalState, elev, floor) {
			elev.SetMoveDir(management.DirDown)
			return
		}
	}

	elev.SetCurrentOrderActiveStatus(false)
}

// TODO: what does this function do?
func assignUp(globalState *state.GlobalState, elev *management.Elevator, startFloor int) bool {

	for floor := startFloor + 1; floor < management.NumFloors; floor++ {

		// CAB priority
		if CabOrderAtFloor(elev, floor) {
			if elev.GetLastOrder().IsHallDownOrder() && HallOrderUpAtFloor(elev, elev.GetFloor()) {
				RemoveHallUp(globalState, elev, elev.GetFloor())
			}
			elev.SetLastOrder(elev.GetCurrentOrder())
			elev.SetCurrentOrder(elev.GetOrder(floor, int(elevIO.CabButton)))
			return true
		}

		// hallUp if going up
		if HallOrderUpAtFloor(elev, floor) && !elev.GetLastOrder().IsHallDownOrder() {
			elev.SetLastOrder(elev.GetCurrentOrder())
			elev.SetCurrentOrder(elev.GetOrder(floor, int(elevIO.HallUpButton)))
			return true
		}

		if !OrdersAbove(elev, floor) && HallOrderDownAtFloor(elev, floor) &&
				(elev.GetLastOrder().IsCabOrder() ||
				(elev.GetLastOrder().IsHallUpOrder() &&
				floor == management.NumFloors-1)) {
			elev.SetLastOrder(elev.GetCurrentOrder())
			elev.SetCurrentOrder(elev.GetOrder(floor, int(elevIO.HallDownButton)))
			return true
		}
	}
	return false
}

// TODO: what does this function do?
func assignDown(globalState *state.GlobalState, elev *management.Elevator, startFloor int) bool {
	for floor := startFloor - 1; floor >= 0; floor-- {
		// CAB prioritet
		if CabOrderAtFloor(elev, floor) {
			if elev.GetLastOrder().IsHallUpOrder() && HallOrderDownAtFloor(elev, elev.GetFloor()) {
				RemoveHallDown(globalState, elev, floor)
			}
			elev.SetLastOrder(elev.GetCurrentOrder())
			elev.SetCurrentOrder(elev.GetOrder(floor, int(elevIO.CabButton)))
			return true
		}

		// hallDown hvis vi går ned
		if HallOrderDownAtFloor(elev, floor) && !elev.GetLastOrder().IsHallUpOrder() {
			elev.SetLastOrder(elev.GetCurrentOrder())
			elev.SetCurrentOrder(elev.GetOrder(floor, int(elevIO.HallDownButton)))
			return true
		}

		// hvis ingen under og hallOrder-logikk oppfylt → kan ta hallUp
		if !OrdersBelow(elev, floor) && HallOrderUpAtFloor(elev, floor) &&
				(elev.GetLastOrder().IsCabOrder() ||
				(elev.GetLastOrder().IsHallDownOrder() && floor == 0)) {
			elev.SetLastOrder(elev.GetCurrentOrder())
			elev.SetCurrentOrder(elev.GetOrder(floor, int(elevIO.HallUpButton)))
			return true
		}
	}
	return false
}

func CabOrderAtFloor(elev *management.Elevator, floor int) bool {
	return elev.GetOrderActiveStatus(floor, int(elevIO.CabButton))
}

func HallOrderUpAtFloor(elev *management.Elevator, floor int) bool {
	return elev.GetOrderActiveStatus(floor, int(elevIO.HallUpButton))
}

func HallOrderDownAtFloor(elev *management.Elevator, floor int) bool {
	return elev.GetOrderActiveStatus(floor, int(elevIO.HallDownButton))
}

// TODO not in use
func AnyOrderAtFloor(elev *management.Elevator, floor int) bool {
	return CabOrderAtFloor(elev, floor) ||
		HallOrderUpAtFloor(elev, floor) ||
		HallOrderDownAtFloor(elev, floor)
}


// TODO remove before handin
func PrintOrders(elev *management.Elevator) {
	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < management.NumButtons; button++ {
			order := elev.GetOrder(floor, button)
			fmt.Printf("Floor: %d Button: %d OrderPlaced: %t\n", order.GetFloor(), order.GetButtonType(), order.GetActiveStatus())
		}
	}
}
