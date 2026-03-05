package orderManagement

import (
	"fmt"
	"heislab/elevio"
	"heislab/management"
)

func ordersAbove(e *management.Elevator, floorUnderInspection int) bool {

	if floorUnderInspection == management.NumFloors-1 {
		return false
	}

	for f := floorUnderInspection + 1; f < management.NumFloors; f++ {
		for b := 0; b < management.NumButtons; b++ {
			if e.Orders[f][b].OrderPlaced {
				return true
			}
		}
	}

	return false
}

func ordersBelow(e *management.Elevator, floorUnderInspection int) bool {

	if floorUnderInspection == 0 {
		return false
	}

	for f := floorUnderInspection - 1; f >= 0; f-- {
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

		if !ordersAbove(e,e.Floor) && hallOrderDownAtFloor(e, floor) {
			return true
		}

	case management.DirDown:

		if cabOrderAtFloor(e, floor) {
			return true
		}

		if hallOrderDownAtFloor(e, floor) {
			return true
		}

		if !ordersBelow(e,e.Floor) && hallOrderUpAtFloor(e, floor) {
			return true
		}

	default:

		if anyOrderAtFloor(e, floor) {
			return true
		}
	}

	return false
}

func ClearOrdersAndTurnOfLights(gs *GlobalState) {

	e := &management.Elev
	floor := e.Floor
	if e.CurrentOrder.Floor == floor {
		e.CurrentOrder.OrderPlaced = false
	}

	// CAB order fjernes alltid
	if e.Orders[floor][elevio.CabButton].OrderPlaced {

		e.Orders[floor][elevio.CabButton].OrderPlaced = false
		gs.UpdateLocalGlobalState()
		elevio.SetButtonLamp(elevio.CabButton, floor, false)
	}

	switch e.MoveDir {
	
	case management.DirUp:

		// fjern hallUp
		if e.Orders[floor][elevio.HallUpButton].OrderPlaced {
			gs.mu.Lock()
			e.Orders[floor][elevio.HallUpButton].OrderPlaced = false
			gs.globalState.HallRequests[floor][elevio.HallUpButton] = false
			gs.mu.Unlock()
			gs.IncrementHallRequestVersion(floor,elevio.HallUpButton)
			elevio.SetButtonLamp(elevio.HallUpButton, floor, false)
		}

	case management.DirDown:
		
		// fjern hallDown
		if e.Orders[floor][elevio.HallDownButton].OrderPlaced {
			gs.mu.Lock()
			e.Orders[floor][elevio.HallDownButton].OrderPlaced = false
			gs.globalState.HallRequests[floor][elevio.HallDownButton] = false
			gs.mu.Unlock()
			gs.IncrementHallRequestVersion(floor,elevio.HallDownButton)
			elevio.SetButtonLamp(elevio.HallDownButton, floor, false)
		}

	default:

		// idle → ta begge? NEI --velger en av dem.. Kan forbedres
		
		if e.Orders[floor][elevio.HallDownButton].OrderPlaced {
			gs.mu.Lock()
			e.Orders[floor][elevio.HallDownButton].OrderPlaced = false
			gs.globalState.HallRequests[floor][elevio.HallDownButton] = false
			gs.mu.Unlock()
			gs.IncrementHallRequestVersion(floor,elevio.HallDownButton)
			elevio.SetButtonLamp(elevio.HallDownButton, floor, false)
		} else if e.Orders[floor][elevio.HallUpButton].OrderPlaced {
			gs.mu.Lock()
			e.Orders[floor][elevio.HallUpButton].OrderPlaced = false
			gs.globalState.HallRequests[floor][elevio.HallUpButton] = false
			gs.mu.Unlock()
			gs.IncrementHallRequestVersion(floor,elevio.HallUpButton)
			elevio.SetButtonLamp(elevio.HallUpButton, floor, false)
		}
	}
}

func UpdateCurrentOrder() {

	e := &management.Elev
	fmt.Println("Entered Update current order with: ")
	fmt.Println("Elev floor:", e.Floor)
	fmt.Println("MoveDir:", e.MoveDir)
	for f := 0; f < management.NumFloors; f++ {
		fmt.Println(
			"floor", f,
			"cab", e.Orders[f][elevio.CabButton].OrderPlaced,
			"up", e.Orders[f][elevio.HallUpButton].OrderPlaced,
			"down", e.Orders[f][elevio.HallDownButton].OrderPlaced,
		)
	}
	floor := e.Floor
	if floor < 0 {
		floor = e.LastFloor
	}

	switch e.MoveDir {

	case management.DirUp:

		if assignUp(e, floor) {
			return
		}

		if ordersBelow(e,e.Floor) {
			assignDown(e, floor)
			e.MoveDir = management.DirDown
			return
		}

	case management.DirDown:

		if assignDown(e, floor) {
			return
		}

		if ordersAbove(e,e.Floor) {
			assignUp(e, floor)
			e.MoveDir = management.DirUp
			return
		}

	default: // Idle

		if assignUp(e, floor) {
			e.MoveDir = management.DirUp
			return
		}
		fmt.Println("Inside default in updateCurrentOrder")
		if assignDown(e, floor) {
			fmt.Println("Assigned a order to currentOrder")
			e.MoveDir = management.DirDown
			return
		}
	}
	fmt.Println("UpdateCurrentOrder did not find any orders")
}

//Find orders upwards
func assignUp(e *management.Elevator, startFloor int) bool {

	for f := startFloor + 1; f < management.NumFloors; f++ {

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
		if !ordersAbove(e,f) && hallOrderDownAtFloor(e, f) {
			e.CurrentOrder = e.Orders[f][elevio.HallDownButton]
			return true
		}
	}

	return false
}

//Find orders downwards
func assignDown(e *management.Elevator, startFloor int) bool {

	for f := startFloor - 1; f >= 0; f-- {

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
		if !ordersBelow(e,f) && hallOrderUpAtFloor(e, f) {
			e.CurrentOrder = e.Orders[f][elevio.HallUpButton]
			return true
		}
	}

	return false
}

func UpdateMoveDir() {
	e := &management.Elev

	if e.CurrentOrder.Floor == -1 {
		e.MoveDir = management.DirIdle
		return
	}

	if e.Floor == -1 {
		// Between floors → keep current direction
		return
	}

	if e.CurrentOrder.Floor > e.Floor {
		e.MoveDir = management.DirUp
	} else if e.CurrentOrder.Floor < e.Floor {
		e.MoveDir = management.DirDown
	} else {
		e.MoveDir = management.DirIdle
	}
}