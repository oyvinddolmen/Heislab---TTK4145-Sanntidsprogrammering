package elevator

import (
	"fmt"
	"heislab/elevio"
	"heislab/management"
	"heislab/network"
	"heislab/orderManagement"
	"time"
)

// Timer for door
var doorTimer *time.Timer
var HallUpAndHallDownAndCabAtDifferentDir bool

const doorOpenDuration = 2 * time.Second

// -------------------------------------------------------------------------------------------
// Initialize FSM
// -------------------------------------------------------------------------------------------

// initializes Elevator struct and creates order matrix
func InitFSM(elevID string, NumFloors int) {
	NoOrder := management.Order{Floor: 1, ButtonType: 2, ElevID: "", OrderPlaced: false}

	management.Elev = management.Elevator{
		ID:           elevID,
		State:        management.ElevInit,
		Floor:        0,
		LastFloor:    0,
		MoveDir:      management.DirIdle,
		CurrentOrder: NoOrder,
		LastOrder:    NoOrder,
	}

	for floor := 0; floor < NumFloors; floor++ {
		for button := 0; button < management.NumButtons; button++ {
			management.Elev.Orders[floor][button] = management.Order{
				Floor:       floor,
				ButtonType:  elevio.ButtonType(button),
				ElevID:      "",
				OrderPlaced: false,
			}
		}
	}
}

// -------------------------------------------------------------------------------------------
// Run FSM
// -------------------------------------------------------------------------------------------

func RunElevator(
	gs *orderManagement.GlobalState,
	elevChannels management.ElevChannels,
	networkChannels network.NetworkConn,
) {
	go elevio.PollFloorSensor(elevChannels.NewFloor)
	go elevio.PollButtons(elevChannels.BtnPresses)
	go elevio.PollStopButton(elevChannels.StopBtn)
	go elevio.PollObstructionSwitch(elevChannels.Obstruction)

	setElevState(gs, management.ElevIdle)
	go runFSM(gs, elevChannels, networkChannels)
}

// -------------------------------------------------------------------------------------------
// FSM loop
// -------------------------------------------------------------------------------------------

func runFSM(
	gs *orderManagement.GlobalState,
	elevChannels management.ElevChannels,
	networkChannels network.NetworkConn,
) {
	for {
		switch management.Elev.State {

		// ----------------- Case: IDLE -------------------------
		case management.ElevIdle:
			select {
			case <-networkChannels.WorldViewUpdate:
				if needToOpenDoors(gs) {
					orderManagement.ServeHallRequestsAtCurrentFloor(gs)
					orderManagement.ClearOrdersAndTurnOfLights(gs)
					SetAllLights(management.Elev, gs)
					setElevState(gs, management.ElevObstruction)
				} else {
					orderManagement.RunHallAssignerAndApplyAssignments(gs)
					SetAllLights(management.Elev, gs)
					UpdateCurrentOrderAndsafeDrive(&management.Elev, gs)
				}
			case <-elevChannels.Obstruction:
				setElevState(gs, management.ElevObstruction)
			case <-elevChannels.StopBtn:
				if management.Elev.Floor != -1 {
					setElevState(gs, management.ElevObstruction)
				} else {
					setElevState(gs, management.ElevStop)
				}
			case btn := <-elevChannels.BtnPresses:
				HallUpAndHallDownAndCabAtDifferentDir = false
				handleButtonPress(gs, btn, networkChannels)
				if getFloor() != -1 {
					HallUpAndHallDownAndCabAtDifferentDir = orderManagement.ClearOrdersAndTurnOfLights(gs)
					orderManagement.RunHallAssignerAndApplyAssignments(gs)
					SetAllLights(management.Elev, gs)
					}
				if HallUpAndHallDownAndCabAtDifferentDir {
					setElevState(gs, management.ElevObstruction)
				} else {UpdateCurrentOrderAndsafeDrive(&management.Elev, gs)}
			}

		// ----------------- Case: MOVING -------------------------
		case management.ElevMoving:
			select {
			case <-networkChannels.WorldViewUpdate:
				orderManagement.RunHallAssignerAndApplyAssignments(gs)
				SetAllLights(management.Elev, gs)
				UpdateCurrentOrderAndsafeDrive(&management.Elev, gs)
			case floor := <-elevChannels.NewFloor:
				setFloorIndicator(floor)
				setElevLastFloor(floor)
				if orderManagement.ShouldStop(&management.Elev, floor) {
					setMotorStop()
					setElevFloor(floor)

					orderManagement.ChooseDirectionAfterStop(&management.Elev, floor)

					orderManagement.ClearOrdersAndTurnOfLights(gs)

					orderManagement.RunHallAssignerAndApplyAssignments(gs)
					if management.Elev.CurrentOrder.OrderPlaced == false {
						orderManagement.UpdateCurrentOrder(gs)
						orderManagement.UpdateMoveDir(&management.Elev)
					}
					setElevState(gs, management.ElevObstruction)
				}
			case <-elevChannels.Obstruction:
				setElevState(gs, management.ElevObstruction)
			case <-elevChannels.StopBtn:
				setElevState(gs, management.ElevStop)
			case btn := <-elevChannels.BtnPresses:
				handleButtonPress(gs, btn, networkChannels)
				orderManagement.UpdateCurrentOrder(gs)
			}

		// ----------------- Case: STOP -------------------------
		case management.ElevStop:
			select {
			case <-networkChannels.WorldViewUpdate:
				orderManagement.RunHallAssignerAndApplyAssignments(gs)
				SetAllLights(management.Elev, gs)
				UpdateCurrentOrderAndsafeDrive(&management.Elev, gs)
			case btn := <-elevChannels.BtnPresses:
				handleButtonPress(gs, btn, networkChannels)
				orderManagement.UpdateCurrentOrder(gs)
			case <-elevChannels.StopBtn:
				if getFloor() != -1 {
					setElevState(gs, management.ElevObstruction)
				} else {
					setElevState(gs, management.ElevIdle)
				}
			}

		// ----------------- Case: OBSTRUCTION -------------------------
		case management.ElevObstruction:
			select {
			case <-networkChannels.WorldViewUpdate:
				if needToOpenDoors(gs) {
					orderManagement.ServeHallRequestsAtCurrentFloor(gs)
					orderManagement.ClearOrdersAndTurnOfLights(gs)
					SetAllLights(management.Elev, gs)
					setElevState(gs, management.ElevObstruction)
				} else {
					orderManagement.RunHallAssignerAndApplyAssignments(gs)
					SetAllLights(management.Elev, gs)
					UpdateCurrentOrderAndsafeDrive(&management.Elev, gs)
				}
			case <-doorTimer.C:
				if !elevio.GetObstruction() {
					elevio.SetDoorOpenLamp(false)
					setElevState(gs, management.ElevIdle)
				} else {
					setElevState(gs, management.ElevObstruction)
				}
			case btn := <-elevChannels.BtnPresses:
				HallUpAndHallDownAndCabAtDifferentDir = false
				handleButtonPress(gs, btn, networkChannels)
				if getFloor() != -1 {
					HallUpAndHallDownAndCabAtDifferentDir = orderManagement.ClearOrdersAndTurnOfLights(gs)
					orderManagement.RunHallAssignerAndApplyAssignments(gs)
					SetAllLights(management.Elev, gs)
					}
				if HallUpAndHallDownAndCabAtDifferentDir {
					doorTimer = time.NewTimer(doorOpenDuration)
				} else {orderManagement.UpdateCurrentOrder(gs)
						orderManagement.UpdateMoveDir(&management.Elev)}
			}
		}
	}
}

// -------------------------------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------------------------------

// creates order, updates global state and runs hallassigner
func handleButtonPress(gs *orderManagement.GlobalState, btn elevio.ButtonEvent, networkChannels network.NetworkConn) {
	order := orderManagement.CreateOrder(btn)

	// Ignore button press if we are already at the floor and serving it, but open door
	if order.Floor == management.Elev.Floor {
		setElevState(gs, management.ElevObstruction)
		return
		// skal vi ikke åpne døren uansett hvis en knapp trykkes i samme etasje som heisen er i? - Øyvind
		// tanken var -> hva hvis noen presser ned, og så kommer noen etter dørene har lukket seg og presser opp... men det er veldig edge case, sikkert unødvendig
		/*
			if order.ButtonType == elevio.CabButton ||
				(order.ButtonType == elevio.HallUpButton && management.Elev.MoveDir == management.DirUp) ||
				(order.ButtonType == elevio.HallDownButton && management.Elev.MoveDir == management.DirDown) {
				setElevState(gs, management.ElevObstruction)
				return
			}
		*/
	}

	if order.ButtonType == management.CabButton {
		orderManagement.AddCabOrderToElevator(order)
		gs.UpdateGlobalState()
	} else {
		gs.AddHallRequest(order)
		gs.IncrementHallRequestVersion(order.Floor, order.ButtonType)
	}

	network.SendGlobalState(gs, networkChannels.GlobalStateTx)
	orderManagement.RunHallAssignerAndApplyAssignments(gs)
	SetAllLights(management.Elev, gs)
}

// sets moveDir in elevator struct
func setMoveDir(moveDir management.Direction) {
	management.Elev.MoveDir = moveDir
}

func setElevLastFloor(lastFloor int) {
	management.Elev.LastFloor = lastFloor
}

func setElevFloor(floor int) {
	management.Elev.Floor = floor
}

// returns elevators floor
func getFloor() int {
	return management.Elev.Floor
}

// sets elevio motordirection to stop
func setMotorStop() {
	elevio.SetMotorDirection(elevio.MotorDirStop)
}

// sets Elev's floor and lastFloor, and sets floor indicator
func updateFloor(floor int) {
	if floor >= 0 {
		management.Elev.Floor = floor
		management.Elev.LastFloor = floor
		elevio.SetFloorIndicator(floor)
	}
}

// updates current order and sets motor direction
func UpdateCurrentOrderAndsafeDrive(e *management.Elevator, gs *orderManagement.GlobalState) {
	orderManagement.UpdateCurrentOrder(gs)
	orderManagement.UpdateMoveDir(e)
	fmt.Println("Calculated moving dir:", management.Elev.MoveDir)
	if management.Elev.MoveDir == management.DirIdle {
		setMotorStop()
		return
	}

	setMotorFromDir()
	setElevState(nil, management.ElevMoving)
}

// checks if there are any hall-orders at the same floor as elevator
func needToOpenDoors(gs *orderManagement.GlobalState) bool {
	currentFloor := management.Elev.Floor
	hallRequests := gs.GetCopy().HallRequests

	if currentFloor == -1 {
		return false
	} else {
		for btn := 0; btn < 2; btn++ {
			if hallRequests[currentFloor][btn] {
				return true
			}
		}
		return false
	}
}

// -------------------------------------------------------------------------------------------
// State transitions
// -------------------------------------------------------------------------------------------

// sets elevators state and call on-state-entry functions
func setElevState(gs *orderManagement.GlobalState, state management.State) {
	prev := management.Elev.State
	management.Elev.State = state

	switch state {
	case management.ElevIdle:
		onIdleEntry(&management.Elev, gs)
	case management.ElevMoving:
		onMovingEntry()
	case management.ElevStop:
		onStopEntry()
	case management.ElevObstruction:
		onObstructionEntry()
	}

	fmt.Println("STATE CHANGE:", prev, "->", state)
}

// turns off door open and stop lamp, and sets motor dir based on next order
func onIdleEntry(e *management.Elevator, gs *orderManagement.GlobalState) {
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	if elevio.GetFloor() == -1 {
		setElevFloor(-1)
	}
	orderManagement.RunHallAssignerAndApplyAssignments(gs)
	UpdateCurrentOrderAndsafeDrive(e, gs)
	SetAllLights(management.Elev, gs)
}

// turns off stop and door open lamp, and sets elevio motor direction
func onMovingEntry() {
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	setElevFloor(-1)
}

// turns on stop lamp and stops elevator
func onStopEntry() {
	elevio.SetStopLamp(true)
	stopElevator()
}

// turns on door open lamp and starts new timer
func onObstructionEntry() {
	stopElevator()
	elevio.SetDoorOpenLamp(true)
	if doorTimer != nil {
		doorTimer.Stop()
	}
	doorTimer = time.NewTimer(doorOpenDuration)
}
