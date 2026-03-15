package elevator

import (
	"fmt"
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/network"
	"heislab/orderManagement"
	"heislab/state"
	"time"
)

// timer variables
var doorTimer *time.Timer
var canTakeOrdersTimer *time.Timer
var idleTimer *time.Timer

// helper variables
var hallUpAndHallDownAndCabAtDifferentDir bool

// time durations
const doorOpenDuration = 2 * time.Second
const canTakeOrdersCountdown = 4 * time.Second
const IdleTimeOut = 2 * time.Second

// -------------------------------------------------------------------------------------------
// Setting state and on-state-entry functions
// -------------------------------------------------------------------------------------------

// sets elevators state and call on-state-entry functions
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

// turns off door open and stop lamp, and sets motor dir based on next order
func onIdleEntry(elev *management.Elevator, globalState *state.GlobalState) {
	elevIO.SetDoorOpenLamp(false)
	elevIO.SetStopLamp(false)
	elev.SetFloor(elevIO.GetFloor())
	updateAssignments(elev, globalState)
	UpdateCurrentOrderAndsafeDrive(elev, globalState)
	SetAllLights(elev, globalState)
	turnOffCanTakeOrdersTimer()
	elev.SetCanTakeOrders(true)
	startIdleTimer()
}

// turns off stop and door open lamp, and sets elevIO motor direction
func onMovingEntry(elev *management.Elevator) {
	elevIO.SetDoorOpenLamp(false)
	elevIO.SetStopLamp(false)
	elev.SetFloor(-1)
	startNewCanTakeOrdersTimer()
}

// turns on door open lamp and starts new timer
func onObstructionEntry(elev *management.Elevator) {
	stopElevator(elev)
	setDoorOpenLampIfNotBetweenFloors()
	startNewDoorTimer()
	startNewCanTakeOrdersTimer()
}

// -------------------------------------------------------------------------------------------
// FSM stopping and moving logic
// -------------------------------------------------------------------------------------------

// creates order, updates global state and runs hallassigner
func handleButtonPress(elev *management.Elevator, globalState *state.GlobalState, button elevIO.ButtonEvent, networkChannels network.NetworkConnection) bool {
	order := management.CreateOrder(button)

	// Ignore button press if we are already at the floor and serving it, but open door
	if order.GetFloor() == elev.GetFloor() {
		setElevState(elev, globalState, management.ElevObstruction)
		return true
	}
	if order.GetButtonType() == management.CabButton {
		elev.SetOrder(order)
		globalState.UpdateGlobalState(elev)
	} else {
		globalState.AddHallOrder(order)
		globalState.IncrementHallOrderVersion(order.GetFloor(), order.GetButtonType())
	}

	network.SendGlobalState(elev, globalState, networkChannels.OutgoingGlobalStateChannel)
	orderManagement.RunHallAssignerAndApplyAssignments(elev, globalState)
	SetAllLights(elev, globalState)
	return false
}

func ShouldStop(elev *management.Elevator, floor int) bool {
	if !elev.GetCurrentOrderActiveStatus() {
		return true
	}

	switch elev.GetMoveDir() {
	case management.DirUp:
		if orderManagement.CabOrderAtFloor(elev, floor) {
			return true
		}

		// Takes hallorderUp when bypassing as long as the current order is not a hallDown not at 4 floor
		if orderManagement.HallOrderUpAtFloor(elev, floor) {
			if elev.GetCurrentOrderButtonType() == elevIO.HallDownButton {
				if elev.GetCurrentOrderFloor() == management.NumFloors-1 {
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
			if elev.GetCurrentOrderButtonType() == elevIO.HallUpButton {
				if elev.GetCurrentOrderFloor() == 0 {
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
func UpdateCurrentOrderAndsafeDrive(elev *management.Elevator, globalState *state.GlobalState) {
	orderManagement.UpdateCurrentOrder(elev, globalState)
	UpdateMoveDir(elev)
	if elev.GetMoveDir() == management.DirIdle {
		setMotorStop()
		return
	}

	setMotorFromDir(elev)
	setElevState(elev, nil, management.ElevMoving)
}

// runs hallAssigner and sets all order lights
func updateAssignments(elev *management.Elevator, globalState *state.GlobalState) {
	orderManagement.RunHallAssignerAndApplyAssignments(elev, globalState)
	SetAllLights(elev, globalState)
}

// checks if there are any hall-orders at the same floor as elevator
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
	if orderManagement.HallOrderUpAtFloor(elev, floor) && elev.GetCurrentOrderButtonType() == elevIO.HallUpButton {
		elev.SetMoveDir(management.DirUp)
		return
	}
	// if the current order was a hallDown -> Set direction to Down
	if orderManagement.HallOrderDownAtFloor(elev, floor) && elev.GetCurrentOrderButtonType() == elevIO.HallDownButton {
		elev.SetMoveDir(management.DirDown)
		return
	}
	UpdateMoveDir(elev)
}

func UpdateMoveDir(elev *management.Elevator) {
	currentOrderFloor := elev.GetCurrentOrderFloor()
	elevFloor := elev.GetFloor()

	if !elev.GetCurrentOrderActiveStatus() {

		fmt.Println("currentorder.orderplaced == false, stop")
		return
	}

	// if between floors, use last floor as reference
	if elevFloor == -1 {
		elevFloor = elev.GetLastFloor()
	}

	fmt.Println("Elev floor used in updateMovDir is:", elevFloor)

	if currentOrderFloor > elevFloor {
		elev.SetMoveDir(management.DirUp)
	} else if currentOrderFloor < elev.GetFloor() {
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
