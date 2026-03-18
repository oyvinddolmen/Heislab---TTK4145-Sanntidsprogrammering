package elevator

import (
	"fmt"
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/orderManagement"
	"heislab/state"
	"time"
)

// Timer variables
var doorTimer *time.Timer
var canTakeOrdersTimer *time.Timer
var idleTimer *time.Timer

// time durations
const doorOpenDuration = 3 * time.Second
const canTakeOrdersCountdown = 4 * time.Second
const IdleTimeOut = 2 * time.Second

// -------------------------------------------------------------------------------------------
// Setting state and on-state-entry functions
// -------------------------------------------------------------------------------------------

// Sets elevators state and calls on-state-entry functions.
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

// Turns off door open and stop lamp, and sets motor dir based on next order.
func onIdleEntry(elev *management.Elevator, globalState *state.GlobalState) {
	fmt.Println("SetterDoorOpenTILFALSE")
	elevIO.SetDoorOpenLamp(false)
	updateAssignments(elev, globalState)
	UpdateCurrentOrderAndDrive(elev, globalState)
	SetAllLights(elev, globalState)
	turnOffCanTakeOrdersTimer()
	elev.SetCanTakeOrders(true)
	startIdleTimer()
}

// Turns off stop and door open lamp, and sets elevIO motor direction.
// TODO: Hvor settes elevIO motor direction ?
func onMovingEntry(elev *management.Elevator) {
	elevIO.SetDoorOpenLamp(false)
	elev.SetFloor(-1)
	startNewCanTakeOrdersTimer()
}

// Stops elevator, turns on door open lamp and starts new timer.
func onObstructionEntry(elev *management.Elevator) {
	stopElevator(elev)
	elevIO.SetDoorOpenLamp(true)
	startNewDoorTimer()
	startNewCanTakeOrdersTimer()
}

// -------------------------------------------------------------------------------------------
// FSM stopping and moving logic
// -------------------------------------------------------------------------------------------

// Creates order, updates global state and runs hall order assigner.
func registerOrder(elev *management.Elevator,
		globalState *state.GlobalState,
		button elevIO.ButtonEvent){
	order := management.CreateOrder(button)
	if order.IsCabOrder() {
		elev.SetOrder(order)
		globalState.UpdateGlobalState(elev)
	} else {
		globalState.AddHallOrder(order)
		globalState.IncrementHallOrderVersion(order.GetFloor(), order.GetButtonType())
	}
}

// Checks if the elevator is currently at a floor with hall up and halldown order, 
// and the button press is in a different direction than the cleared hallorder
func isCabOrderAtDifferentDir(elev *management.Elevator) bool {
	if elev.GetMoveDir() != management.DirIdle{
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

func atButtonFloor(elev *management.Elevator, button elevIO.ButtonEvent) bool {
	if button.Floor == elev.GetFloor() {
		return true
	}
	return false
}

// Determines if elevator should stop at current floor.
func ShouldStop(elev *management.Elevator, floor int) bool {
	if !elev.GetCurrentOrderActiveStatus() || orderManagement.CabOrderAtFloor(elev, floor) {
		return true
	}
	if elev.GetCurrentOrderFloor() == floor{
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

// Updates current order and sets motor direction.
func UpdateCurrentOrderAndDrive(elev *management.Elevator, globalState *state.GlobalState) {
	updateCurrentOrderAndMoveDir(elev, globalState)
	setMotorFromDir(elev)
	if elev.GetMoveDir() != management.DirIdle{
		setElevState(elev, nil, management.ElevMoving)
	}
}

// Runs hall order assigner and sets all order lights.
func updateAssignments(elev *management.Elevator, globalState *state.GlobalState) {
	orderManagement.RunHallAssignerAndApplyAssignments(elev, globalState)
	SetAllLights(elev, globalState)
}

func updateCurrentOrderAndMoveDir(elev *management.Elevator, globalState *state.GlobalState) {
	orderManagement.UpdateCurrentOrder(elev, globalState)
	UpdateMoveDir(elev)
}



// Checks if there are any hall orders at the same floor as elevator.
func needToOpenDoors(elev *management.Elevator, globalState *state.GlobalState) bool {
	currentFloor := elev.GetFloor()
	if currentFloor == -1 {
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

func UpdateMoveDir(elev *management.Elevator) {
	currentOrderFloor := elev.GetCurrentOrderFloor()
	elevFloor := elev.GetFloor()

	// If elev has no assigned orders.
	if !elev.GetCurrentOrderActiveStatus() {
		// TODO remove print
		fmt.Println("currentorder.orderplaced == false, stop")
		elev.SetMoveDir(management.DirIdle)
		return
	}
	if elevFloor == -1 {
		elevFloor = elev.GetLastFloor()
	}

	if currentOrderFloor > elevFloor {
		elev.SetMoveDir(management.DirUp)
	} else if currentOrderFloor < elevFloor {
		elev.SetMoveDir(management.DirDown)
	} else {
		elev.SetMoveDir(management.DirIdle)
	}
}

// -------------------------------------------------------------------------------------------
// Timer functions
// -------------------------------------------------------------------------------------------

func startNewDoorTimer() {
	if doorTimer != nil {
		doorTimer.Stop()
	}
	doorTimer = time.NewTimer(doorOpenDuration)
}

func startNewCanTakeOrdersTimer() {
	if canTakeOrdersTimer != nil {
		canTakeOrdersTimer.Stop()
	}
	canTakeOrdersTimer = time.NewTimer(canTakeOrdersCountdown)
	fmt.Println("Started a new canTakeOrdersTimer ----------")
}

func turnOffCanTakeOrdersTimer() {
	if canTakeOrdersTimer != nil {
		fmt.Println("Turned off canTakeOrdersTimer ----------")
		canTakeOrdersTimer.Stop()
	}
}

func resetCanTakeOrdersTimer() {
	if canTakeOrdersTimer != nil {
		canTakeOrdersTimer.Reset(canTakeOrdersCountdown)
	}
	fmt.Println("Reset canTakeOrderTimer ---------------")
}

func startIdleTimer() {
	if idleTimer != nil {
		idleTimer.Stop()
	}
	idleTimer = time.NewTimer(IdleTimeOut)
	fmt.Println("Started Idle Timer ---------------")

}
