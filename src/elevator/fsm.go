package elevator

import (
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/network"
	"heislab/orderManagement"
	"heislab/state"
	"time"
	"fmt"
)

type ElevChannels struct {
	NewFloor    chan int
	Obstruction chan bool
	StopBtn     chan bool
	BtnPresses  chan elevIO.ButtonEvent 
}

var doorTimer *time.Timer
var idleTimer *time.Timer

const doorOpenDuration = 2 * time.Second
const IdleTimeOut = 2 * time.Second


func RunElevator(
	elev *management.Elevator,
	gs *state.GlobalState,
	elevChannels ElevChannels,
	networkChannels network.NetworkConn,
) {
	go elevIO.PollFloorSensor(elevChannels.NewFloor)
	go elevIO.PollButtons(elevChannels.BtnPresses)
	go elevIO.PollStopButton(elevChannels.StopBtn)
	go elevIO.PollObstructionSwitch(elevChannels.Obstruction)
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
					orderManagement.ClearOrdersAndTurnOffLights(elev, gs)
					SetAllLights(elev, gs)
					setElevState(elev, gs, management.ElevObstruction)
				} else {
					updateAssignments(elev, gs)
					UpdateCurrentOrderAndsafeDrive(elev, gs)
				}
			case <-idleTimer.C: //If been Idle too long, we make it so the elevator 
								// takes all orders no matter if lastorder was a hallUp or a hallDown
				elev.LastOrder.ButtonType = elevIO.CabButton 
				UpdateCurrentOrderAndsafeDrive(elev, gs)
				if elev.CurrentOrder.OrderPlaced == false {
					resetTimer(&idleTimer, IdleTimeOut)
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
				HallUpAndHallDownAndCabAtDifferentDir := false
				OrderWasAtCurrentFloor := handleButtonPress(elev, gs, btn, networkChannels)
				if !OrderWasAtCurrentFloor {
					if elev.GetFloor() != -1 {
						HallUpAndHallDownAndCabAtDifferentDir = orderManagement.ClearOrdersAndTurnOffLights(elev, gs)
						updateAssignments(elev, gs)
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
				updateAssignments(elev, gs)
			case floor := <-elevChannels.NewFloor:
				setFloorIndicator(floor)
				elev.SetElevLastFloor(floor)
				if ShouldStop(elev, floor) {
					setMotorStop()
					elev.SetElevFloor(floor)
					ChooseDirectionAfterStop(elev, floor)
					orderManagement.ClearOrdersAndTurnOffLights(elev, gs)
					updateAssignments(elev, gs)
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
					orderManagement.ClearOrdersAndTurnOffLights(elev, gs)
					SetAllLights(elev, gs)
					resetTimer(&doorTimer, doorOpenDuration)
				} else {
					updateAssignments(elev, gs)
					orderManagement.UpdateCurrentOrder(elev, gs)
					UpdateMoveDir(elev)
				}
			case <-doorTimer.C:
				if !elevIO.GetObstruction() {
					setElevState(elev, gs, management.ElevIdle)
				} else {
					setElevState(elev, gs, management.ElevObstruction)
				}
			case btn := <-elevChannels.BtnPresses:
				mixedHallOrders := false
				OrderWasAtCurrentFloor := handleButtonPress(elev, gs, btn, networkChannels)
				if !OrderWasAtCurrentFloor {
					
					if elev.GetFloor() != -1 {
						mixedHallOrders = orderManagement.ClearOrdersAndTurnOffLights(elev, gs)
						updateAssignments(elev, gs)
						}
					if mixedHallOrders || needToOpenDoors(elev, gs){
						doorTimer = time.NewTimer(doorOpenDuration)
					} else {orderManagement.UpdateCurrentOrder(elev, gs)
							UpdateMoveDir(elev)}
					}
			}
		}
	}
}

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

// turns off door open and stop lamp, and sets motor dir based on next order
func onIdleEntry(elev *management.Elevator, gs *state.GlobalState) {
	elevIO.SetDoorOpenLamp(false)
	elevIO.SetStopLamp(false)
	if elevIO.GetFloor() == -1 {
		elev.SetElevFloor(-1)
	}
	updateAssignments(elev, gs)
	UpdateCurrentOrderAndsafeDrive(elev, gs)
	resetTimer(&idleTimer, IdleTimeOut)
}

// turns off stop and door open lamp, and sets elevIO motor direction
func onMovingEntry(elev *management.Elevator) {
	elevIO.SetDoorOpenLamp(false)
	elevIO.SetStopLamp(false)
	elev.SetElevFloor(-1)
}

// turns on stop lamp and stops elevator
func onStopEntry(elev *management.Elevator) {
	elevIO.SetStopLamp(true)
	stopElevator(elev)
}

// turns on door open lamp and starts new timer
func onObstructionEntry(elev *management.Elevator) {
	stopElevator(elev)
	if elev.GetFloor() != -1{
		elevIO.SetDoorOpenLamp(true)
	}
	resetTimer(&doorTimer, doorOpenDuration)
}

func resetTimer(t **time.Timer, duration time.Duration) {
	if *t != nil {
		if !(*t).Stop() {
			select {
			case <-(*t).C:
			default:
			}
		}
	}
	*t = time.NewTimer(duration)
}