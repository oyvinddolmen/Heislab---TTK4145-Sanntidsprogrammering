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

// checks if elev should stop at floor
func ShouldStop(e *management.Elevator, floor int) bool {

	switch e.MoveDir {

	case management.DirUp:

		if cabOrderAtFloor(e, floor) {
			return true
		}

		//Takes hallorderUp when bypassing as long as the current order is not a hallDown not at 4 floor 
		if hallOrderUpAtFloor(e, floor){
			if (e.CurrentOrder.ButtonType == elevio.HallDownButton){
				if e.CurrentOrder.Floor == management.NumFloors{
					return true
				}
			} else {return true}
		}

		if !ordersAbove(e, floor) && hallOrderDownAtFloor(e, floor) {
			return true
		}
		if floor == management.NumFloors-1 {
			return true
		}

	case management.DirDown:

		if cabOrderAtFloor(e, floor) {
			return true
		}

		//Takes hallorderUp when bypassing as long as the current order is not a hallDown not at 4 floor 
		if hallOrderDownAtFloor(e, floor){
			if (e.CurrentOrder.ButtonType == elevio.HallUpButton){
				if e.CurrentOrder.Floor == 0 {
					return true
				}
			} else {return true}
		}

		if !ordersBelow(e, floor) && hallOrderUpAtFloor(e, floor) {
			return true
		}
		if floor == 0 {
			return true
		}
	}

	return false
}

func ClearOrdersAndTurnOfLights(gs *GlobalState) {

	e := &management.Elev
	fmt.Println("Entered ClearOrdersAndTurnOfLights order with: ")
	fmt.Println("Elev floor:", e.Floor)
	fmt.Println("MoveDir:", e.MoveDir)
	fmt.Println("LastOrder.ButtonType (0: HallUp , 1: HallDown , 2: Caborder) : ", e.LastOrder.ButtonType)
	for f := 0; f < management.NumFloors; f++ {
		fmt.Println(
			"floor", f,
			"cab", e.Orders[f][elevio.CabButton].OrderPlaced,
			"up", e.Orders[f][elevio.HallUpButton].OrderPlaced,
			"down", e.Orders[f][elevio.HallDownButton].OrderPlaced,
		)
	}
	floor := e.Floor
	if e.CurrentOrder.Floor == floor {
		e.CurrentOrder.OrderPlaced = false
		e.LastOrder = e.CurrentOrder
	}

	// CAB orders are always removed
	if e.Orders[floor][elevio.CabButton].OrderPlaced {
		removeCabOrder(gs, e, floor)
	}

	switch e.MoveDir {

	case management.DirUp:

		// fjern hallUp
		if e.Orders[floor][elevio.HallUpButton].OrderPlaced {
			removeHallUp(gs, e, floor)
		}

	case management.DirDown:

		// fjern hallDown
		if e.Orders[floor][elevio.HallDownButton].OrderPlaced {
			removeHallDown(gs, e, floor)
		}

	default:

	if e.Orders[floor][elevio.HallUpButton].OrderPlaced &&
		e.LastOrder.ButtonType == elevio.HallDownButton &&
		cabOrderAbove(e, floor) {
		fmt.Println("Going up!") //Dont remove this print - Elevator is supposed alert
		removeHallUp(gs, e, floor)

	} else if e.Orders[floor][elevio.HallDownButton].OrderPlaced &&
				e.LastOrder.ButtonType == elevio.HallUpButton &&
				cabOrderBelow(e, floor) {
					fmt.Println("Going Down!") //Dont remove this print - Elevator is supposed alert
					removeHallDown(gs, e, floor)

				}
	}
}

func cabOrderAbove(e *management.Elevator, floor int) bool {
	for f := floor + 1; f < management.NumFloors; f++ {
		if e.Orders[f][elevio.CabButton].OrderPlaced {
			return true
		}
	}
	return false
}

func cabOrderBelow(e *management.Elevator, floor int) bool {
	for f := floor - 1; f >= 0; f-- {
		if e.Orders[f][elevio.CabButton].OrderPlaced {
			return true
		}
	}
	return false
}
func removeCabOrder(gs *GlobalState, e *management.Elevator, floor int) {
	e.Orders[floor][elevio.CabButton].OrderPlaced = false
	gs.UpdateGlobalState()
	elevio.SetButtonLamp(elevio.CabButton, floor, false)
}

func removeHallDown(gs *GlobalState, e *management.Elevator, floor int) {
	gs.mu.Lock()
	e.Orders[floor][elevio.HallDownButton].OrderPlaced = false
	gs.globalState.HallRequests[floor][elevio.HallDownButton] = false
	gs.mu.Unlock()
	gs.IncrementHallRequestVersion(floor, elevio.HallDownButton)
	elevio.SetButtonLamp(elevio.HallDownButton, floor, false)
}

func removeHallUp(gs *GlobalState, e *management.Elevator, floor int) {
	gs.mu.Lock()
	e.Orders[floor][elevio.HallUpButton].OrderPlaced = false
	gs.globalState.HallRequests[floor][elevio.HallUpButton] = false
	gs.mu.Unlock()
	gs.IncrementHallRequestVersion(floor, elevio.HallUpButton)
	elevio.SetButtonLamp(elevio.HallUpButton, floor, false)
}

func UpdateCurrentOrder(gs *GlobalState) {
	e := &management.Elev
	
	fmt.Println("Entered Update current order with: ")
	fmt.Println("Elev floor:", e.Floor)
	fmt.Println("MoveDir:", e.MoveDir)
	fmt.Println("LastOrder.ButtonType (0: HallUp , 1: HallDown , 2: Caborder) : ", e.LastOrder.ButtonType)
	for f := 0; f < management.NumFloors; f++ {
		fmt.Println(
			"floor", f,
			"cab", e.Orders[f][elevio.CabButton].OrderPlaced,
			"up", e.Orders[f][elevio.HallUpButton].OrderPlaced,
			"down", e.Orders[f][elevio.HallDownButton].OrderPlaced,
		)
	}
	
	floor := e.Floor
	if e.Floor == -1 {
		return
	}

	switch e.MoveDir {

	case management.DirUp:

		if assignUp(gs, e, floor) {
			return
		}

		if ordersBelow(e, e.Floor) {
			assignDown(gs, e, floor)
			e.MoveDir = management.DirDown
			return
		}

	case management.DirDown:

		if assignDown(gs, e, floor) {
			return
		}

		if ordersAbove(e, e.Floor) {
			assignUp(gs, e, floor)
			e.MoveDir = management.DirUp
			return
		}

	default: // Idle

		if assignUp(gs, e, floor) {
			e.MoveDir = management.DirUp
			return
		}
		if assignDown(gs, e, floor) {
			e.MoveDir = management.DirDown
			return
		}
	}
	//fmt.Println("UpdateCurrentOrder did not find any orders")
}

// Find orders upwards
func assignUp(gs *GlobalState, e *management.Elevator, startFloor int) bool {

	for f := startFloor + 1; f < management.NumFloors; f++ {

		// CAB prioritet
		if cabOrderAtFloor(e, f) {
			if e.LastOrder.ButtonType == elevio.HallDownButton && hallOrderUpAtFloor(e, e.Floor) {
				removeHallUp(gs, e, e.Floor)
			}
			e.LastOrder = e.CurrentOrder
			e.CurrentOrder = e.Orders[f][elevio.CabButton]
			return true
		}

		// hallUp hvis vi går opp
		if hallOrderUpAtFloor(e, f) && e.LastOrder.ButtonType != elevio.HallDownButton {
			e.LastOrder = e.CurrentOrder
			e.CurrentOrder = e.Orders[f][elevio.HallUpButton]
			return true
		}

		// hvis ingen over og hallOrdreLogikk oppfylt→ kan ta hallDown
		if !ordersAbove(e, f) && hallOrderDownAtFloor(e, f) && (e.LastOrder.ButtonType == elevio.CabButton || (e.LastOrder.ButtonType == elevio.HallUpButton && f == management.NumFloors - 1)) {
			e.LastOrder = e.CurrentOrder
			e.CurrentOrder = e.Orders[f][elevio.HallDownButton]
			return true
		}
	}

	return false
}

// Find orders downwards
func assignDown(gs *GlobalState, e *management.Elevator, startFloor int) bool {

	for f := startFloor - 1; f >= 0; f-- {

		// CAB prioritet
		if cabOrderAtFloor(e, f) {
			if e.LastOrder.ButtonType == elevio.HallUpButton && hallOrderDownAtFloor(e, e.Floor) {
				removeHallDown(gs, e, e.Floor)
			}
			e.LastOrder = e.CurrentOrder
			e.CurrentOrder = e.Orders[f][elevio.CabButton]
			return true
		}

		// hallDown hvis vi går ned
		if hallOrderDownAtFloor(e, f) && e.LastOrder.ButtonType != elevio.HallUpButton {
			e.LastOrder = e.CurrentOrder
			e.CurrentOrder = e.Orders[f][elevio.HallDownButton]
			return true
		}

		// hvis ingen under og hallorderLogikk oppfyllt→ kan ta hallUp 
		if !ordersBelow(e, f) && hallOrderUpAtFloor(e, f) && (e.LastOrder.ButtonType == elevio.CabButton || (e.LastOrder.ButtonType == elevio.HallDownButton && f == 0)) {
			e.LastOrder = e.CurrentOrder
			e.CurrentOrder = e.Orders[f][elevio.HallUpButton]
			return true	
		}
	}

	return false
}

// updates elevator-struct's moveDir
func UpdateMoveDir(e *management.Elevator) {

	if e.Floor == -1 || e.CurrentOrder.OrderPlaced == false {
		// Between floors or no currenOrder(this means we are initializing) → keep current direction
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

func ChooseDirectionAfterStop(e *management.Elevator, floor int) {
	// if the currentorder was a hallUP -> Set direction to Up
	if hallOrderUpAtFloor(e, floor) && e.CurrentOrder.ButtonType == elevio.HallUpButton {
		e.MoveDir = management.DirUp
		return
	}
	// if the currentorder was a hallDown -> Set direction to down
	if hallOrderDownAtFloor(e, floor) && e.CurrentOrder.ButtonType == elevio.HallDownButton {
		e.MoveDir = management.DirDown
		return
	}

	UpdateMoveDir(e)
}
