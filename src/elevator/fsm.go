package elevator

import (
	"fmt"
	"heislab/elevio"
	"heislab/management"
	"heislab/network"
	"heislab/orderManagement"
	"time"
)

// -------------------------------------------------------------------------------------------
// Initialize state-machine
// -------------------------------------------------------------------------------------------

func InitFSM(elevID string, NumFloors int) {
	noOrder := management.Order{Floor: -1, ButtonType: -1, ElevID: ""}

	management.Elev.ID = elevID
	management.Elev.State = management.ElevInit
	management.Elev.Floor = -1
	management.Elev.LastFloor = 0
	management.Elev.MoveDir = management.DirDown
	management.Elev.CurrentOrder = noOrder

	for floor := 0; floor < NumFloors; floor++ {
		for button := 0; button < management.NumButtons; button++ {
			management.Elev.Orders[floor][button] = management.Order{
				Floor:      floor,
				ButtonType: elevio.ButtonType(button),
				ElevID:     "",
				//Finished:    false,
				OrderPlaced: false,
			}
		}
	}
}

// -------------------------------------------------------------------------------------------
// Timer for door
// -------------------------------------------------------------------------------------------

var doorTimer *time.Timer

const doorOpenDuration = 2 * time.Second

// -------------------------------------------------------------------------------------------
// Run Elevator
// -------------------------------------------------------------------------------------------

func RunElevator(
	gs *orderManagement.GlobalState,
	elevChannels management.ElevChannels,
	networkChannels network.NetworkConn,
) {
	go elevio.PollFloorSensor(elevChannels.LastFloor)
	go elevio.PollButtons(elevChannels.BtnPresses)
	go elevio.PollStopButton(elevChannels.StopBtn)
	go elevio.PollObstructionSwitch(elevChannels.Obstruction)

	go runFSM(gs, elevChannels, networkChannels)
}

// -------------------------------------------------------------------------------------------
// FSM
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
				// TODO: make function handleWorldViewUpdate(gs)
				setHallLightOnAllPanels(gs)
				orderManagement.RunHallAssigner(gs)
				motorDir := findMotorDirection(
					management.Elev.CurrentOrder.Floor,
					management.Elev.LastFloor,
				)
				setMoveDir(management.Direction(motorDir))
				setMotorFromMoveDir()
				setStateFromMoveDir(gs)

			case <-elevChannels.Obstruction:
				setElevState(gs, management.ElevObstruction)

			case <-elevChannels.StopBtn:
				if management.Elev.Floor != -1 {
					setElevState(gs, management.ElevObstruction)
				} else {
					setElevState(gs, management.ElevStop)
				}

			case btnPress := <-elevChannels.BtnPresses:

				// TODO: create function handleBtnPressIdle(order, gs)

				order := orderManagement.CreateOrder(btnPress)
				if order.Floor == management.Elev.Floor {
					setElevState(gs, management.ElevObstruction)
					continue

				}
				if order.ButtonType == management.CabButton {
					orderManagement.AddOrderToOrders(order)
					gs.UpdateLocalGlobalState()
				} else {
					gs.AddHallRequest(order)
					gs.IncrementHallRequestVersion(order)
				}

				network.SendGlobalState(gs, networkChannels.GlobalStateTx)
				fmt.Println("Valid order floor", order.Floor, "btn:", btnPress.Button)
				elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)

				// finding next order and set ElevMoving if elev has new order
				orderManagement.RunHallAssigner(gs)
				motorDir := findMotorDirection(
					management.Elev.CurrentOrder.Floor,
					management.Elev.LastFloor,
				)
				setMoveDir(management.Direction(motorDir))
				setMotorFromMoveDir()
				setStateFromMoveDir(gs)
			}

		// ----------------- Case: MOVING -------------------------
		case management.ElevMoving:
			select {

			case <-networkChannels.WorldViewUpdate:
				// handleWorldViewUpdate
				setHallLightOnAllPanels(gs)
				orderManagement.RunHallAssigner(gs)
				motorDir := findMotorDirection(
					management.Elev.CurrentOrder.Floor,
					management.Elev.LastFloor,
				)
				setMoveDir(management.Direction(motorDir))
				setMotorFromMoveDir()
				setStateFromMoveDir(gs)

			case <-elevChannels.StopBtn:
				setElevState(gs, management.ElevStop)

			case floor := <-elevChannels.LastFloor:
				setElevLastFloor(floor)
				elevio.SetFloorIndicator(floor)

				if reachedDestination(floor) {
					stopElevator()
					orderManagement.CompleteCurrentOrder(gs)
					reachedFloorLightsOff(floor)
					setElevFloor(floor)
					setElevState(gs, management.ElevObstruction)
				}

			case btnPress := <-elevChannels.BtnPresses:

				// TODO: make handleBtnPressMoving(btn) function

				order := orderManagement.CreateOrder(btnPress)

				if order.ButtonType == management.CabButton {
					orderManagement.AddOrderToOrders(order)
					gs.UpdateLocalGlobalState()
				} else {
					gs.AddHallRequest(order)
					gs.IncrementHallRequestVersion(order)
				}

				network.SendGlobalState(gs, networkChannels.GlobalStateTx)
				elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)
				fmt.Println("Valid order floor", order.Floor, "btn:", btnPress.Button)

				orderManagement.RunHallAssigner(gs)
				motorDir := findMotorDirection(
					management.Elev.CurrentOrder.Floor,
					management.Elev.LastFloor,
				)
				setMoveDir(management.Direction(motorDir))
				setMotorFromMoveDir()
				setStateFromMoveDir(gs)
			}

		// ----------------- Case: STOP -------------------------
		case management.ElevStop:
			select {

			case <-networkChannels.WorldViewUpdate:
				setHallLightOnAllPanels(gs)
				orderManagement.RunHallAssigner(gs)

			case btnPress := <-elevChannels.BtnPresses:

				// TODO: handleBtnPressStop()

				order := orderManagement.CreateOrder(btnPress)

				if order.ButtonType == management.CabButton {
					orderManagement.AddOrderToOrders(order)
					gs.UpdateLocalGlobalState()
				} else {
					gs.AddHallRequest(order)
					gs.IncrementHallRequestVersion(order)
				}

				network.SendGlobalState(gs, networkChannels.GlobalStateTx)
				orderManagement.RunHallAssigner(gs)
				elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)
				fmt.Println("Valid order floor", order.Floor, "btn:", btnPress.Button)

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
				setHallLightOnAllPanels(gs)
				orderManagement.RunHallAssigner(gs)

			case <-doorTimer.C:
				if !elevio.GetObstruction() {
					elevio.SetDoorOpenLamp(false)
					setElevState(gs, management.ElevIdle)
				}

			case obstructed := <-elevChannels.Obstruction:
				if !obstructed {
					doorTimer.Reset(doorOpenDuration)
				}

			case btnPress := <-elevChannels.BtnPresses:

				// TODO: handleBtnPressObstruction()
				order := orderManagement.CreateOrder(btnPress)

				if order.Floor == getFloor() {
					setElevState(gs, management.ElevObstruction)
					continue

				} else {
					if order.ButtonType == management.CabButton {
						orderManagement.AddOrderToOrders(order)
						gs.UpdateLocalGlobalState()
					} else {
						gs.AddHallRequest(order)
						gs.IncrementHallRequestVersion(order)
					}
				}

				network.SendGlobalState(gs, networkChannels.GlobalStateTx)

				fmt.Println("Valid order floor", order.Floor, "btn:", btnPress.Button)
				elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)
			}
		}
	}
}

// -------------------------------------------------------------------------------------------
// State transitions
// -------------------------------------------------------------------------------------------

func onStopEntry() {
	elevio.SetStopLamp(true)
	setMotorStop()
}

// turns off doorOpen and stop lamp and calls setMotorFromMoveDir
func onMovingEntry() {
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	setMotorFromMoveDir()
}

// turns off stop and obstruction lights and calculates new moving dir and sets it
func onIdleEntry(gs *orderManagement.GlobalState) {
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	setMoveDir(management.DirIdle)
	elevio.SetMotorDirection(elevio.MotorDirStop)

	orderManagement.RunHallAssigner(gs)
	motorDir := findMotorDirection(
		management.Elev.CurrentOrder.Floor,
		management.Elev.LastFloor,
	)
	setMoveDir(management.Direction(motorDir))
	setMotorFromMoveDir()
	if motorDir != elevio.MotorDirStop {
		setStateFromMoveDir(gs)
	}
}

func onObstructionEntry() {
	setMotorStop()
	elevio.SetDoorOpenLamp(true)

	if doorTimer != nil {
		doorTimer.Stop()
	}
	doorTimer = time.NewTimer(doorOpenDuration)
}

// -------------------------------------------------------------------------------------------
// Get and set functions
// -------------------------------------------------------------------------------------------

func setElevState(gs *orderManagement.GlobalState, state management.State) {
	prev := management.Elev.State
	management.Elev.State = state

	switch state {
	case management.ElevIdle:
		onIdleEntry(gs)

	case management.ElevMoving:
		onMovingEntry()

	case management.ElevStop:
		onStopEntry()

	case management.ElevObstruction:
		onObstructionEntry()
	}

	fmt.Println("STATE CHANGE:", prev, "->", state)
}

// sets the FSM state based on moving direction in Elev struct
func setStateFromMoveDir(gs *orderManagement.GlobalState) {
	switch management.Elev.MoveDir {
	case management.DirDown:
		setElevState(gs, management.ElevMoving)

	case management.DirUp:
		setElevState(gs, management.ElevMoving)

	case management.DirIdle:
		setElevState(gs, management.ElevIdle)
	}

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

// returns elevators last floor
func getLastFloor() int {
	return management.Elev.LastFloor
}

// -------------------------------------------------------------------------------------------
// Motor helpers
// -------------------------------------------------------------------------------------------

func setMotorStop() {
	elevio.SetMotorDirection(elevio.MotorDirStop)
}

func setMotorFromMoveDir() {
	switch management.Elev.MoveDir {
	case management.DirUp:
		elevio.SetMotorDirection(elevio.MotorDirUp)
	case management.DirDown:
		elevio.SetMotorDirection(elevio.MotorDirDown)
	default:
		elevio.SetMotorDirection(elevio.MotorDirStop)
	}
}
