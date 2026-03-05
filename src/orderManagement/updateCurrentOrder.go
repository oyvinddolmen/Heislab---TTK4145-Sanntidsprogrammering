package orderManagement

import (
	"heislab/elevio"
	"heislab/management"
)

func ordersAbove(e *management.Elevator) bool {
	for f := e.Floor + 1; f < management.NumFloors; f++ {
		for b := 0; b < management.NumButtons; b++ {
			if e.Orders[f][b].OrderPlaced {
				return true
			}
		}
	}
	return false
}

func ordersBelow(e *management.Elevator) bool {
	for f := e.Floor - 1; f >= 0; f-- {
		for b := 0; b < management.NumButtons; b++ {
			if e.Orders[f][b].OrderPlaced {
				return true
			}
		}
	}
	return false
}

func cabOrderAtFloor(e *management.Elevator, floor int) bool {
	return e.Orders[floor][elevio.CabButton].OrderPlaced
}

func hallOrderUpAtFloor(e *management.Elevator, floor int) bool {
	return e.Orders[floor][elevio.HallUpButton].OrderPlaced
}

func hallOrderDownAtFloor(e *management.Elevator, floor int) bool {
	return e.Orders[floor][elevio.HallDownButton].OrderPlaced
}

func anyOrderAtFloor(e *management.Elevator, floor int) bool {
	return cabOrderAtFloor(e, floor) ||
		hallOrderUpAtFloor(e, floor) ||
		hallOrderDownAtFloor(e, floor)
}

func ShouldStop(e *management.Elevator) bool {

	floor := e.Floor

	switch e.MoveDir {

	case management.DirUp:

		if cabOrderAtFloor(e, floor) {
			return true
		}

		if hallOrderUpAtFloor(e, floor) {
			return true
		}

		if !ordersAbove(e) && hallOrderDownAtFloor(e, floor) {
			return true
		}

	case management.DirDown:

		if cabOrderAtFloor(e, floor) {
			return true
		}

		if hallOrderDownAtFloor(e, floor) {
			return true
		}

		if !ordersBelow(e) && hallOrderUpAtFloor(e, floor) {
			return true
		}

	default:

		if anyOrderAtFloor(e, floor) {
			return true
		}
	}

	return false
}

func ClearOrdersAtFloor(gs *GlobalState) {

	e := &management.Elev
	floor := e.Floor

	// CAB order fjernes alltid
	if e.Orders[floor][elevio.CabButton].OrderPlaced {

		e.Orders[floor][elevio.CabButton].OrderPlaced = false
		gs.UpdateLocalGlobalState()
	}

	gs.mu.Lock()

	switch e.MoveDir {

	case management.DirUp:

		// fjern hallUp
		if e.Orders[floor][elevio.HallUpButton].OrderPlaced {

			e.Orders[floor][elevio.HallUpButton].OrderPlaced = false
			gs.globalState.HallRequests[floor][elevio.HallUpButton] = false
		}

		// hvis ingen ordre over → ta også hallDown
		if !ordersAbove(e) && e.Orders[floor][elevio.HallDownButton].OrderPlaced {

			e.Orders[floor][elevio.HallDownButton].OrderPlaced = false
			gs.globalState.HallRequests[floor][elevio.HallDownButton] = false
		}

	case management.DirDown:

		// fjern hallDown
		if e.Orders[floor][elevio.HallDownButton].OrderPlaced {

			e.Orders[floor][elevio.HallDownButton].OrderPlaced = false
			gs.globalState.HallRequests[floor][elevio.HallDownButton] = false
		}

		// hvis ingen ordre under → ta også hallUp
		if !ordersBelow(e) && e.Orders[floor][elevio.HallUpButton].OrderPlaced {

			e.Orders[floor][elevio.HallUpButton].OrderPlaced = false
			gs.globalState.HallRequests[floor][elevio.HallUpButton] = false
		}

	default:

		// idle → ta begge
		if e.Orders[floor][elevio.HallUpButton].OrderPlaced {

			e.Orders[floor][elevio.HallUpButton].OrderPlaced = false
			gs.globalState.HallRequests[floor][elevio.HallUpButton] = false
		}

		if e.Orders[floor][elevio.HallDownButton].OrderPlaced {

			e.Orders[floor][elevio.HallDownButton].OrderPlaced = false
			gs.globalState.HallRequests[floor][elevio.HallDownButton] = false
		}
	}

	gs.mu.Unlock()

	UpdateCurrentOrder()
}

func UpdateCurrentOrder() {

	e := &management.Elev

	//If already working an order - dont switch
	//if e.CurrentOrder.OrderPlaced {
	//	return
	//}

	floor := e.Floor
	if floor < 0 {
		floor = e.LastFloor
	}

	switch e.MoveDir {

	case management.DirUp:

		if assignUp(e, floor) {
			return
		}

		if ordersBelow(e) {
			e.MoveDir = management.DirDown
			assignDown(e, floor)
			return
		}

	case management.DirDown:

		if assignDown(e, floor) {
			return
		}

		if ordersAbove(e) {
			e.MoveDir = management.DirUp
			assignUp(e, floor)
			return
		}

	default: // Idle

		if assignUp(e, floor) {
			e.MoveDir = management.DirUp
			return
		}

		if assignDown(e, floor) {
			e.MoveDir = management.DirDown
			return
		}

		e.State = management.ElevIdle
	}
}

//Find orders upwards
func assignUp(e *management.Elevator, startFloor int) bool {

	for f := startFloor; f < management.NumFloors; f++ {

		// CAB prioritet
		if cabOrderAtFloor(e, f) {
			e.CurrentOrder = e.Orders[f][elevio.CabButton]
			return true
		}

		// hallUp hvis vi går opp
		if hallOrderUpAtFloor(e, f) {
			e.CurrentOrder = e.Orders[f][elevio.HallUpButton]
			return true
		}

		// hvis ingen over → kan ta hallDown
		if !ordersAbove(e) && hallOrderDownAtFloor(e, f) {
			e.CurrentOrder = e.Orders[f][elevio.HallDownButton]
			return true
		}
	}

	return false
}

//Find orders downwards
func assignDown(e *management.Elevator, startFloor int) bool {

	for f := startFloor; f >= 0; f-- {

		// CAB prioritet
		if cabOrderAtFloor(e, f) {
			e.CurrentOrder = e.Orders[f][elevio.CabButton]
			return true
		}

		// hallDown hvis vi går ned
		if hallOrderDownAtFloor(e, f) {
			e.CurrentOrder = e.Orders[f][elevio.HallDownButton]
			return true
		}

		// hvis ingen under → kan ta hallUp
		if !ordersBelow(e) && hallOrderUpAtFloor(e, f) {
			e.CurrentOrder = e.Orders[f][elevio.HallUpButton]
			return true
		}
	}

	return false
}


// removes currentorder from globalState and Elev
func CompleteCurrentOrder(gs *GlobalState) {

	e := &management.Elev
	floor := e.CurrentOrder.Floor
	button := e.CurrentOrder.ButtonType

	// CAB ORDER
	if button == elevio.CabButton {

		e.Orders[floor][button].OrderPlaced = false
		e.CurrentOrder.OrderPlaced = false

		gs.UpdateLocalGlobalState()

	} else {

		// HALL ORDER

		e.CurrentOrder.OrderPlaced = false

		gs.mu.Lock()

		// Fjern bare ordre i samme retning
		if e.MoveDir == management.DirUp {
			gs.globalState.HallRequests[floor][elevio.HallUpButton] = false
		}

		if e.MoveDir == management.DirDown {
			gs.globalState.HallRequests[floor][elevio.HallDownButton] = false
		}

		gs.mu.Unlock()

		gs.IncrementHallRequestVersion(e.CurrentOrder)

		// Hvis cab også fantes i etasjen → fjern den
		if e.Orders[floor][elevio.CabButton].OrderPlaced {
			e.Orders[floor][elevio.CabButton].OrderPlaced = false
			gs.UpdateLocalGlobalState()
		}
	}

	UpdateCurrentOrder()
}