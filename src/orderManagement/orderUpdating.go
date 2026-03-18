package orderManagement

import (
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/state"
)

// Selects the next active order based on the elevator's current floor and travel direction
func UpdateCurrentOrder(elev *management.Elevator, globalState *state.GlobalState) {
	floor := elev.GetFloor()
	if !elev.IsAtFloor() {
		return
	}

	switch elev.GetMoveDir() {
	case management.DirUp:
		if assignCurrentOrderAbove(globalState, elev, floor) {
			return
		}

		if OrdersBelow(elev, elev.GetFloor()) {
			assignCurrentOrderBelow(globalState, elev, floor)
			elev.SetMoveDir(management.DirDown)
			return
		}

	case management.DirDown:
		if assignCurrentOrderBelow(globalState, elev, floor) {
			return
		}

		if OrdersAbove(elev, elev.GetFloor()) {
			assignCurrentOrderAbove(globalState, elev, floor)
			elev.SetMoveDir(management.DirUp)
			return
		}

	default: // Idle
		if assignCurrentOrderAbove(globalState, elev, floor) {
			elev.SetMoveDir(management.DirUp)
			return
		}
		if assignCurrentOrderBelow(globalState, elev, floor) {
			elev.SetMoveDir(management.DirDown)
			return
		}
	}

	elev.SetCurrentOrderActiveStatus(false)
}

// Scans floors above startFloor and assigns the next upward-reachable order.
// Returns true if it sets new current order.
func assignCurrentOrderAbove(globalState *state.GlobalState, elev *management.Elevator, startFloor int) bool {

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

		// hallUp orders if going up
		if HallOrderUpAtFloor(elev, floor) && !elev.GetLastOrder().IsHallDownOrder() {
			elev.SetLastOrder(elev.GetCurrentOrder())
			elev.SetCurrentOrder(elev.GetOrder(floor, int(elevIO.HallUpButton)))
			return true
		}

		// if no orders above -> can take hallDown orders
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

// Scans floors below startFloor and assigns the next downward-reachable order.
// Returns true if it sets new current order.
func assignCurrentOrderBelow(globalState *state.GlobalState, elev *management.Elevator, startFloor int) bool {
	for floor := startFloor - 1; floor >= 0; floor-- {
		// CAB priority
		if CabOrderAtFloor(elev, floor) {
			if elev.GetLastOrder().IsHallUpOrder() && HallOrderDownAtFloor(elev, elev.GetFloor()) {
				RemoveHallDown(globalState, elev, floor)
			}
			elev.SetLastOrder(elev.GetCurrentOrder())
			elev.SetCurrentOrder(elev.GetOrder(floor, int(elevIO.CabButton)))
			return true
		}

		// hallDown orders if going down
		if HallOrderDownAtFloor(elev, floor) && !elev.GetLastOrder().IsHallUpOrder() {
			elev.SetLastOrder(elev.GetCurrentOrder())
			elev.SetCurrentOrder(elev.GetOrder(floor, int(elevIO.HallDownButton)))
			return true
		}

		// If no orders below -> can take hallUp orders
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
