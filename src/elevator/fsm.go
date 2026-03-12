package elevator

import (
	"fmt"
	"heislab/elevator/elevio"
	"heislab/management"
	"heislab/network"
	"heislab/orderManagement"
	"heislab/state"
	"time"
)

// Timer for door
var doorTimer *time.Timer
var IdleTimer *time.Timer
var HallUpAndHallDownAndCabAtDifferentDir bool
var OrderWasAtCurrentFloor bool

const doorOpenDuration = 2 * time.Second
const IdleTimeOut = 2 * time.Second

func InitElevator(elevID string, numFloors int) management.Elevator {
    noOrder := management.Order{
        Floor: 1,
        ButtonType: elevio.CabButton,
        OrderPlaced: false,
    }
    elev := management.Elevator{
        ID:           elevID,
        State:        management.ElevInit,
        Floor:        0,
        LastFloor:    0,
        MoveDir:      management.DirIdle,
        CurrentOrder: noOrder,
        LastOrder:    noOrder,
    }
    for floor := 0; floor < numFloors; floor++ {
        for button := 0; button < management.NumButtons; button++ {
            elev.Orders[floor][button] = management.Order{
                Floor: floor,
                ButtonType: elevio.ButtonType(button),
                OrderPlaced: false,
            }
        }
    }
    return elev
}

func RunElevator(
	elev *management.Elevator,
	gs *state.GlobalState,
	elevChannels ElevChannels,
	networkChannels network.NetworkConn,
) {
	go elevio.PollFloorSensor(elevChannels.NewFloor)
	go elevio.PollButtons(elevChannels.BtnPresses)
	go elevio.PollStopButton(elevChannels.StopBtn)
	go elevio.PollObstructionSwitch(elevChannels.Obstruction)
	setElevState(elev, gs, management.ElevIdle)
	go runFSM(elev, gs, elevChannels, networkChannels)
}

// -------------------------------------------------------------------------------------------
// FSM loop
// -------------------------------------------------------------------------------------------

func runFSM(
	elev *management.Elevator,
	gs *state.GlobalState,
	elevChannels ElevChannels,
	networkChannels network.NetworkConn,
) {
	for {
		switch elev.State {

		// ----------------- Case: IDLE -------------------------
		case management.ElevIdle:
			select {
			case <-networkChannels.WorldViewUpdate: 
				if needToOpenDoors(elev, gs) {
					orderManagement.ServeHallRequestsAtCurrentFloor(elev, gs)
					orderManagement.ClearOrdersAndTurnOfLights(elev, gs)
					SetAllLights(elev, gs)
					setElevState(elev, gs, management.ElevObstruction)
				} else {
					orderManagement.RunHallAssignerAndApplyAssignments(elev, gs)
					SetAllLights(elev, gs)
					UpdateCurrentOrderAndsafeDrive(elev, gs)
				}
			case <-IdleTimer.C: //If been Idle too long, we make it so the elevator 
								// takes all orders no matter if lastorder was a hallUp or a hallDown
				elev.LastOrder.ButtonType = elevio.CabButton 
				UpdateCurrentOrderAndsafeDrive(elev, gs)
				if elev.CurrentOrder.OrderPlaced == false {
					doorTimer = time.NewTimer(doorOpenDuration)
				}
			case <-elevChannels.Obstruction:
				setElevState(elev, gs, management.ElevObstruction)
			case <-elevChannels.StopBtn:
				if elev.Floor != -1 {
					setElevState(elev, gs, management.ElevObstruction)
				} else {
					setElevState(elev, gs, management.ElevStop)
				}
			case btn := <-elevChannels.BtnPresses:
				HallUpAndHallDownAndCabAtDifferentDir = false
				OrderWasAtCurrentFloor = handleButtonPress(elev, gs, btn, networkChannels)
				if !OrderWasAtCurrentFloor {
					if elev.GetFloor() != -1 {
						HallUpAndHallDownAndCabAtDifferentDir = orderManagement.ClearOrdersAndTurnOfLights(elev, gs)
						orderManagement.RunHallAssignerAndApplyAssignments(elev, gs)
						SetAllLights(elev, gs)
						}
					if HallUpAndHallDownAndCabAtDifferentDir{
						setElevState(elev, gs, management.ElevObstruction)
					} else {UpdateCurrentOrderAndsafeDrive(elev, gs)}
				}
			}

		// ----------------- Case: MOVING -------------------------
		case management.ElevMoving:
			select {
			case <-networkChannels.WorldViewUpdate:
				orderManagement.RunHallAssignerAndApplyAssignments(elev, gs)
				SetAllLights(elev, gs)
			case floor := <-elevChannels.NewFloor:
				setFloorIndicator(floor)
				elev.SetElevLastFloor(floor)
				if ShouldStop(elev, floor) {
					setMotorStop()
					elev.SetElevFloor(floor)
					ChooseDirectionAfterStop(elev, floor)
					orderManagement.ClearOrdersAndTurnOfLights(elev, gs)
					orderManagement.RunHallAssignerAndApplyAssignments(elev, gs)
					if elev.CurrentOrder.OrderPlaced == false {
						orderManagement.UpdateCurrentOrder(elev, gs)
						UpdateMoveDir(elev)
					}
					setElevState(elev, gs, management.ElevObstruction)
				}
			case <-elevChannels.Obstruction:
				setElevState(elev, gs, management.ElevObstruction)
			case <-elevChannels.StopBtn:
				setElevState(elev, gs, management.ElevStop)
			case btn := <-elevChannels.BtnPresses:
				handleButtonPress(elev, gs, btn, networkChannels)
				orderManagement.UpdateCurrentOrder(elev, gs)
			}

		// ----------------- Case: OBSTRUCTION -------------------------
		case management.ElevObstruction:
			select {
			case <-networkChannels.WorldViewUpdate:
				if needToOpenDoors(elev, gs) {
					orderManagement.ServeHallRequestsAtCurrentFloor(elev, gs)
					orderManagement.ClearOrdersAndTurnOfLights(elev, gs)
					SetAllLights(elev, gs)
					doorTimer = time.NewTimer(doorOpenDuration)
				} else {
					orderManagement.RunHallAssignerAndApplyAssignments(elev, gs)
					SetAllLights(elev, gs)
					orderManagement.UpdateCurrentOrder(elev, gs)
					UpdateMoveDir(elev)
				}
			case <-doorTimer.C:
				if !elevio.GetObstruction() {
					elevio.SetDoorOpenLamp(false)
					setElevState(elev, gs, management.ElevIdle)
				} else {
					setElevState(elev, gs, management.ElevObstruction)
				}
			case btn := <-elevChannels.BtnPresses:
				HallUpAndHallDownAndCabAtDifferentDir = false
				OrderWasAtCurrentFloor = handleButtonPress(elev, gs, btn, networkChannels)
				if !OrderWasAtCurrentFloor {
					
					if elev.GetFloor() != -1 {
						HallUpAndHallDownAndCabAtDifferentDir = orderManagement.ClearOrdersAndTurnOfLights(elev, gs)
						orderManagement.RunHallAssignerAndApplyAssignments(elev, gs)
						SetAllLights(elev, gs)
						}
					if HallUpAndHallDownAndCabAtDifferentDir || needToOpenDoors(elev, gs){
						doorTimer = time.NewTimer(doorOpenDuration)
					} else {orderManagement.UpdateCurrentOrder(elev, gs)
							UpdateMoveDir(elev)}
					}
			}
		}
	}
}

// -------------------------------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------------------------------

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

// sets Elev's floor and lastFloor, and sets floor indicator
func updateFloor(floor int) {
	if floor >= 0 {
		//elev.Floor = floor
		//elev.LastFloor = floor
		elevio.SetFloorIndicator(floor) //---------------------------
	}
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

// -------------------------------------------------------------------------------------------
// State transitions
// -------------------------------------------------------------------------------------------

// sets elevators state and call on-state-entry functions
func setElevState(elev *management.Elevator, gs *state.GlobalState, state management.State) {
	prev := elev.State
	elev.State = state

	switch state {
	case management.ElevIdle:
		onIdleEntry(elev, gs)
	case management.ElevMoving:
		onMovingEntry(elev)
	case management.ElevStop:
		onStopEntry(elev)
	case management.ElevObstruction:
		onObstructionEntry(elev)
	}

	fmt.Println("STATE CHANGE:", prev, "->", state)
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

// turns off door open and stop lamp, and sets motor dir based on next order
func onIdleEntry(elev *management.Elevator, gs *state.GlobalState) {
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	if elevio.GetFloor() == -1 {
		elev.SetElevFloor(-1)
	}
	orderManagement.RunHallAssignerAndApplyAssignments(elev, gs)
	UpdateCurrentOrderAndsafeDrive(elev, gs)
	SetAllLights(elev, gs)
	if IdleTimer != nil {
		IdleTimer.Stop()
	}
	IdleTimer = time.NewTimer(IdleTimeOut)
}

// turns off stop and door open lamp, and sets elevio motor direction
func onMovingEntry(elev *management.Elevator) {
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	elev.SetElevFloor(-1)
}

// turns on stop lamp and stops elevator
func onStopEntry(elev *management.Elevator) {
	elevio.SetStopLamp(true)
	stopElevator(elev)
}

// turns on door open lamp and starts new timer
func onObstructionEntry(elev *management.Elevator) {
	stopElevator(elev)
	if elev.GetFloor() != 0{
		elevio.SetDoorOpenLamp(true)
	}
	if doorTimer != nil {
		doorTimer.Stop()
	}
	doorTimer = time.NewTimer(doorOpenDuration)
}
