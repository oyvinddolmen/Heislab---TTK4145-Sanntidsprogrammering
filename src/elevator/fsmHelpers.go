package elevator

import (
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/network"
	"heislab/orderManagement"
	"heislab/state"
)

// -------------------------------------------------------------------------------------------
// Setting state and on-state-entry functions
// -------------------------------------------------------------------------------------------

// Sets elevators state and calls on-state-entry functions
func setElevState(elev *management.Elevator, globalState *state.GlobalState, newState management.State) {
	elev.SetState(newState)

	switch newState {
	case management.ElevIdle:
		onIdleEntry(elev, globalState)
	case management.ElevMoving:
		onMovingEntry(elev)
	case management.ElevObstruction:
		onObstructionEntry(elev)
	}
}

// Turns off door open lamp, finds and drives to next order
func onIdleEntry(elev *management.Elevator, globalState *state.GlobalState) {
	elevIO.SetDoorOpenLamp(false)
	updateCurrentOrderAndDrive(elev, globalState)
	turnOffCanTakeOrdersTimer()
	elev.SetCanTakeOrders(true)
	if !elev.GetCurrentOrderActiveStatus() {
		startIdleTimer()
	}
}

func onMovingEntry(elev *management.Elevator) {
	elevIO.SetDoorOpenLamp(false)
	elev.SetFloor(-1)
	startNewCanTakeOrdersTimer()
}

// Stops elevator and opens door
func onObstructionEntry(elev *management.Elevator) {
	stopElevator(elev)
	elevIO.SetDoorOpenLamp(true)
	startNewDoorTimer()
	startNewCanTakeOrdersTimer()
}

// -------------------------------------------------------------------------------------------
// FSM stopping and moving logic
// -------------------------------------------------------------------------------------------

// Checks if the elevator is currently at a floor with hall up and halldown order,
// and the button press is in a different direction than the cleared hallorder
func isCabOrderAtDifferentDir(elev *management.Elevator) bool {
	if elev.GetMoveDir() != management.DirIdle {
		return false
	}
	currentFloor := elev.GetFloor()
	if elev.GetOrderActiveStatus(currentFloor, int(elevIO.HallUpButton)) &&
		elev.GetLastOrder().IsHallDownOrder() &&
		orderManagement.CabOrderAbove(elev, currentFloor) {
		return true

	} else if elev.GetOrderActiveStatus(currentFloor, int(elevIO.HallDownButton)) &&
		elev.GetLastOrder().IsHallUpOrder() &&
		orderManagement.CabOrderBelow(elev, currentFloor) {
		return true
	}
	return false
}

// returns true if elevator is at button's floor
func atButtonFloor(elev *management.Elevator, button elevIO.ButtonEvent) bool {
	if button.Floor == elev.GetFloor() {
		return true
	}
	return false
}

// called when elevator reaches a new floor
func shouldStop(elev *management.Elevator, floor int) bool {
	if !elev.GetCurrentOrderActiveStatus() || orderManagement.CabOrderAtFloor(elev, floor) {
		return true
	}
	if elev.GetCurrentOrderFloor() == floor {
		return true
	}

	switch elev.GetMoveDir() {
	case management.DirUp:
		// Takes HallUp order when bypassing as long as current order is not a HallDown at floor 1-3.
		if orderManagement.HallOrderUpAtFloor(elev, floor) {
			currentOrderIsHallDown := elev.GetCurrentOrder().IsHallDownOrder()
			currentOrderAtTopFloor := elev.GetCurrentOrderFloor() == management.NumFloors-1

			if !currentOrderIsHallDown || currentOrderAtTopFloor {
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
		// Takes HallDown order when bypassing as long as current order is not a HallUp at floor 2-4.
		if orderManagement.HallOrderDownAtFloor(elev, floor) {
			currentOrderIsHallUp := elev.GetCurrentOrder().IsHallUpOrder()
			currentOrderAtBottomFloor := elev.GetCurrentOrderFloor() == 0
			if !currentOrderIsHallUp || currentOrderAtBottomFloor {
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

func updateCurrentOrderAndDrive(elev *management.Elevator, globalState *state.GlobalState) {
	updateCurrentOrderAndMoveDir(elev, globalState)
	setMotorFromDir(elev)
	if elev.GetMoveDir() != management.DirIdle {
		setElevState(elev, nil, management.ElevMoving)
	}
}

// Checks if there are any hall orders at the same floor as elevator.
func hallOrderAtFloor(elev *management.Elevator, globalState *state.GlobalState) bool {
	currentFloor := elev.GetFloor()
	if !elev.IsAtFloor() {
		return false
	}

	hallOrders := globalState.GetCopy().HallOrders

	for button := 0; button < management.NumHallButtonTypes; button++ {
		if hallOrders[currentFloor][button] {
			return true
		}
	}
	return false
}

func ChooseDirectionAfterStop(elev *management.Elevator, floor int) {
	// if the currentorder was a hallUp -> Set direction to Up
	if elev.GetCurrentOrder().IsHallUpOrder() && orderManagement.HallOrderUpAtFloor(elev, floor) {
		elev.SetMoveDir(management.DirUp)
		return
	}
	// if the current order was a hallDown -> Set direction to Down
	if elev.GetCurrentOrder().IsHallDownOrder() && orderManagement.HallOrderDownAtFloor(elev, floor) {
		elev.SetMoveDir(management.DirDown)
		return
	}
	UpdateMoveDir(elev)
}

// finds and updates elevators moving direction
func UpdateMoveDir(elev *management.Elevator) {
	// If elev has no assigned orders.
	if !elev.GetCurrentOrderActiveStatus() {
		elev.SetMoveDir(management.DirIdle)
		return
	}

	currentOrderFloor := elev.GetCurrentOrderFloor()
	currentFloor := elev.GetFloor()
	if !elev.IsAtFloor() {
		currentFloor = elev.GetLastFloor()
	}

	if currentOrderFloor > currentFloor {
		elev.SetMoveDir(management.DirUp)
	} else if currentOrderFloor < currentFloor {
		elev.SetMoveDir(management.DirDown)
	} else {
		elev.SetMoveDir(management.DirIdle)
	}
}

// -------------------------------------------------------------------------------------------
// FSM Readbility functions
// -------------------------------------------------------------------------------------------

func updateAssignmentsAndSetLights(elev *management.Elevator, globalState *state.GlobalState) {
	orderManagement.RunHallAssignerAndApplyAssignments(elev, globalState)
	SetAllLights(elev, globalState)
}

func updateCurrentOrderAndMoveDir(elev *management.Elevator, globalState *state.GlobalState) {
	orderManagement.UpdateCurrentOrder(elev, globalState)
	UpdateMoveDir(elev)
}

func registerAndBroadcastOrder(
	elev *management.Elevator,
	globalState *state.GlobalState,
	button elevIO.ButtonEvent,
	networkChannels network.NetworkChannels,
) {
	order := management.CreateOrder(button)
	if order.IsCabOrder() {
		elev.SetOrder(order)
		globalState.UpdateGlobalState(elev)
	} else {
		globalState.AddHallOrder(order)
		globalState.IncrementHallOrderVersion(order.GetFloor(), order.GetButtonType())
	}
	network.SendGlobalState(elev, globalState, networkChannels.OutgoingGlobalStateChannel)
}

func clearOrdersAndBroadcast(
	elev *management.Elevator,
	globalState *state.GlobalState,
	networkChannels network.NetworkChannels,
) {
	orderManagement.ClearOrdersAtCurrentFloor(elev, globalState)
	network.SendGlobalState(elev, globalState, networkChannels.OutgoingGlobalStateChannel)
}
