package orderManagement

import (
	"fmt"
	"heislab/elevator/elevio"
	"heislab/management"
	"heislab/state"
)

// updates current order based on elev's moving direction
func UpdateCurrentOrder(elev *management.Elevator, gs *state.GlobalState) {
	elev.PrintOrdersDebug()
	floor := elev.Floor
	if elev.Floor == -1 {
		return
	}

	switch elev.MoveDir {
	case management.DirUp:
		if assignUp(gs, elev, floor) {
			return
		}

		if OrdersBelow(elev, elev.Floor) {
			assignDown(gs, elev, floor)
			elev.SetMoveDir(management.DirDown)
			return
		}

	case management.DirDown:
		if assignDown(gs, elev, floor) {
			return
		}

		if OrdersAbove(elev, elev.Floor) {
			assignUp(gs, elev, floor)
			elev.SetMoveDir(management.DirUp)
			return
		}

	default: // Idle
		if assignUp(gs, elev, floor) {
			elev.SetMoveDir(management.DirUp)
			return
		}
		if assignDown(gs, elev, floor) {
			elev.SetMoveDir(management.DirDown)
			return
		}
	}
	//fmt.Println("UpdateCurrentOrder did not find any orders")
	elev.CurrentOrder.OrderPlaced = false
}

// Find orders upwards
func assignUp(gs *state.GlobalState, elev *management.Elevator, startFloor int) bool {

	for floor := startFloor + 1; floor < management.NumFloors; floor++ {

		// CAB prioritet
		if CabOrderAtFloor(elev, floor) {
			if elev.LastOrder.ButtonType == elevio.HallDownButton && HallOrderUpAtFloor(elev, elev.Floor) {
				RemoveHallUp(gs, elev, elev.Floor)
			}
			elev.LastOrder = elev.CurrentOrder
			elev.CurrentOrder = elev.Orders[floor][elevio.CabButton]
			return true
		}

		// hallUp hvis vi går opp
		if HallOrderUpAtFloor(elev, floor) && elev.LastOrder.ButtonType != elevio.HallDownButton {
			elev.LastOrder = elev.CurrentOrder
			elev.CurrentOrder = elev.Orders[floor][elevio.HallUpButton]
			return true
		}

		// hvis ingen over og hallOrdreLogikk oppfylt→ kan ta hallDown
		if !OrdersAbove(elev, floor) && HallOrderDownAtFloor(elev, floor) && 
					(elev.LastOrder.ButtonType == elevio.CabButton || (elev.LastOrder.ButtonType == elevio.HallUpButton && floor == management.NumFloors-1)) {
			elev.LastOrder = elev.CurrentOrder
			elev.CurrentOrder = elev.Orders[floor][elevio.HallDownButton]
			return true
		}
	}

	return false
}

func assignDown(gs *state.GlobalState, elev *management.Elevator, startFloor int) bool {
	for floor := startFloor - 1; floor >= 0; floor-- {
		// CAB prioritet
		if CabOrderAtFloor(elev, floor) {
			if elev.LastOrder.ButtonType == elevio.HallUpButton && HallOrderDownAtFloor(elev, elev.Floor) {
				RemoveHallDown(gs, elev, floor)
			}
			elev.LastOrder = elev.CurrentOrder
			elev.CurrentOrder = elev.Orders[floor][elevio.CabButton]
			return true
		}

		// hallDown hvis vi går ned
		if HallOrderDownAtFloor(elev, floor) && elev.LastOrder.ButtonType != elevio.HallUpButton {
			elev.LastOrder = elev.CurrentOrder
			elev.CurrentOrder = elev.Orders[floor][elevio.HallDownButton]
			return true
		}

		// hvis ingen under og hallOrder-logikk oppfylt → kan ta hallUp
		if !OrdersBelow(elev, floor) && HallOrderUpAtFloor(elev, floor) &&
			(elev.LastOrder.ButtonType == elevio.CabButton ||
				(elev.LastOrder.ButtonType == elevio.HallDownButton && floor == 0)) {
			elev.LastOrder = elev.CurrentOrder
			elev.CurrentOrder = elev.Orders[floor][elevio.HallUpButton]
			return true
		}
	}
	return false
}

func CabOrderAtFloor(elev *management.Elevator, floor int) bool {
	return elev.Orders[floor][elevio.CabButton].OrderPlaced
}

func HallOrderUpAtFloor(elev *management.Elevator, floor int) bool {
	return elev.Orders[floor][elevio.HallUpButton].OrderPlaced
}

func HallOrderDownAtFloor(elev *management.Elevator, floor int) bool {
	return elev.Orders[floor][elevio.HallDownButton].OrderPlaced
}

func AnyOrderAtFloor(elev *management.Elevator, floor int) bool {
	return CabOrderAtFloor(elev, floor) ||
		HallOrderUpAtFloor(elev, floor) ||
		HallOrderDownAtFloor(elev, floor)
}

func PrintOrders(elev *management.Elevator) {
	for f := 0; f < management.NumFloors; f++ {
		for b := 0; b < management.NumButtons; b++ {
			order := elev.Orders[f][b]
			fmt.Printf("Floor: %d Button: %d OrderPlaced: %t\n", order.Floor, order.ButtonType, order.OrderPlaced)
		}
	}
}