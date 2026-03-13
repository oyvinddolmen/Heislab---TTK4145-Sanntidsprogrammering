package elevator

import (
	"fmt"
	"heislab/elevator/elevio"
	"heislab/management"
	"heislab/network"
	"heislab/orderManagement"
	"heislab/state"
)

// creates order, updates global state and runs hallassigner
func handleButtonPress(elev *management.Elevator, gs *state.GlobalState, btn elevio.ButtonEvent, networkChannels network.NetworkConn) bool{
	order := management.CreateOrder(btn)

	// Ignore button press if we are already at the floor and serving it, but open door
	if order.Floor == elev.GetFloor() {
		setElevState(elev, gs, management.ElevObstruction)
		return true
	}
	if order.ButtonType == management.CabButton {
		elev.AddCabOrderToElevator(order)
		gs.UpdateGlobalState(elev)
	} else {
		gs.AddHallRequest(order)
		gs.IncrementHallRequestVersion(order.Floor, order.ButtonType)
	}

	network.SendGlobalState(elev, gs, networkChannels.GlobalStateTx)
	orderManagement.RunHallAssignerAndApplyAssignments(elev, gs)
	SetAllLights(elev, gs)
	return false
}

func ShouldStop(elev *management.Elevator, floor int) bool {
	if elev.CurrentOrder.OrderPlaced == false{
		return true
	}

	switch elev.MoveDir {
	case management.DirUp:
		if orderManagement.CabOrderAtFloor(elev, floor) {
			return true
		}

		// Takes hallorderUp when bypassing as long as the current order is not a hallDown not at 4 floor
		if orderManagement.HallOrderUpAtFloor(elev, floor) {
			if elev.CurrentOrder.ButtonType == elevio.HallDownButton {
				if elev.CurrentOrder.Floor == management.NumFloors-1 {
					return true
				}
			} else {
				return true
			}
		}

		if !orderManagement.OrdersAbove(elev, floor) && orderManagement.HallOrderDownAtFloor(elev, floor) {
			return true
		}
		if floor == management.NumFloors-1 {
			return true
		}

	case management.DirDown:
		if orderManagement.CabOrderAtFloor(elev, floor) {
			return true
		}

		//Takes hallorderUp when bypassing as long as the current order is not a hallDown not at 4 floor
		if orderManagement.HallOrderDownAtFloor(elev, floor) {
			if elev.CurrentOrder.ButtonType == elevio.HallUpButton {
				if elev.CurrentOrder.Floor == 0 {
					return true
				}
			} else {
				return true
			}
		}

		if !orderManagement.OrdersBelow(elev, floor) && orderManagement.HallOrderUpAtFloor(elev, floor) {
			return true
		}
		if floor == 0 {
			return true
		}
	}

	return false
}

// updates current order and sets motor direction
func UpdateCurrentOrderAndsafeDrive(elev *management.Elevator, gs *state.GlobalState) {
	orderManagement.UpdateCurrentOrder(elev, gs)
	UpdateMoveDir(elev)
	if elev.MoveDir == management.DirIdle {
		setMotorStop()
		return
	}

	setMotorFromDir(elev)
	setElevState(elev, nil, management.ElevMoving)
}

func updateAssignments(elev *management.Elevator, gs *state.GlobalState) {
	orderManagement.RunHallAssignerAndApplyAssignments(elev, gs)
	SetAllLights(elev, gs)
}

// checks if there are any hall-orders at the same floor as elevator
func needToOpenDoors(elev *management.Elevator, gs *state.GlobalState) bool {

	currentFloor := elev.Floor
	if currentFloor == -1 {
		return false
	}

	hallRequests := gs.GetCopy().HallRequests
	
	for btn := 0; btn < 2; btn++ {
		if hallRequests[currentFloor][btn] {
			return true
		}
	}
	return false
}

func ChooseDirectionAfterStop(elev *management.Elevator, floor int) {
	// if the currentorder was a hallUP -> Set direction to Up
	if orderManagement.HallOrderUpAtFloor(elev, floor) && elev.CurrentOrder.ButtonType == elevio.HallUpButton {
		elev.MoveDir = management.DirUp
		return
	}
	// if the currentorder was a hallDown -> Set direction to down
	if orderManagement.HallOrderDownAtFloor(elev, floor) && elev.CurrentOrder.ButtonType == elevio.HallDownButton {
		elev.MoveDir = management.DirDown
		return
	}
	UpdateMoveDir(elev)
}

func UpdateMoveDir(elev *management.Elevator) {
	currentOrderFloor := elev.CurrentOrder.Floor
	elevFloor := elev.Floor

	if elev.CurrentOrder.OrderPlaced == false {
		
		fmt.Println("currentorder.orderplaced == false, stop")
		return
	}

	// if between floors, use last floor as reference
	if elevFloor == -1 {
		elevFloor = elev.LastFloor
	}

	fmt.Println("Elev floor used in updateMovDir is:", elevFloor)

	if currentOrderFloor > elevFloor {
		elev.MoveDir = management.DirUp
	} else if currentOrderFloor < elev.Floor {
		elev.MoveDir = management.DirDown
	} else {
		elev.MoveDir = management.DirIdle
	}
}