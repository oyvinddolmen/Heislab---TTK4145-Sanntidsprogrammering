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
				Floor:       floor,
				ButtonType:  elevio.ButtonType(button),
				ElevID:      "",
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

		// =========================================================
		case management.ElevIdle:

			select {

			case <-networkChannels.WorldViewUpdate:
				setHallLightOnAllPanels(gs)
				orderManagement.RunHallAssigner(gs)
				driveToDestination(
					management.Elev.CurrentOrder.Floor,
					management.Elev.LastFloor,
				)
				if management.Elev.MoveDir != management.DirIdle {
					setElevState(management.ElevMoving)
				}
			
			case <-elevChannels.Obstruction:
				setElevState(management.ElevObstruction)
			
			case <-elevChannels.StopBtn:
				if management.Elev.Floor != -1 {
					setElevState(management.ElevObstruction)
				} else {
					setElevState(management.ElevStop)
				}

			case btnPress := <-elevChannels.BtnPresses:

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

				driveToDestination(
					management.Elev.CurrentOrder.Floor,
					management.Elev.LastFloor,
				)

				if management.Elev.MoveDir != management.DirIdle {
					setElevState(management.ElevMoving)
				}

			}

		// =========================================================
		case management.ElevMoving:

			select {
			
			case <-networkChannels.WorldViewUpdate:
				setHallLightOnAllPanels(gs)
				orderManagement.RunHallAssigner(gs)
				driveToDestination(
					management.Elev.CurrentOrder.Floor,
					management.Elev.LastFloor,
				)
				if management.Elev.MoveDir != management.DirIdle {
					setElevState(management.ElevMoving)
				}
			
			case <-elevChannels.StopBtn:
				setElevState(management.ElevStop)
				
			case floor := <-elevChannels.LastFloor:

				management.Elev.LastFloor = floor
				elevio.SetFloorIndicator(floor)

				if reachedDestination(floor) {
					stopElevator()
					orderManagement.CompleteCurrentOrder(gs)
					reachedFloorLightsOff(floor)
					management.Elev.Floor = floor
					setElevState(management.ElevObstruction)
				}

			case btnPress := <-elevChannels.BtnPresses:

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

				driveToDestination(
					management.Elev.CurrentOrder.Floor,
					management.Elev.LastFloor,
				)
			}

		// =========================================================
		case management.ElevStop:

			select {

			case <-networkChannels.WorldViewUpdate:
				setHallLightOnAllPanels(gs)
				orderManagement.RunHallAssigner(gs)

			case btnPress := <-elevChannels.BtnPresses:

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

				if management.Elev.MoveDir != management.DirIdle {
					setElevState(management.ElevMoving)
				}
				
			case <-elevChannels.StopBtn:
				setElevState(management.ElevIdle)
			}

		case management.ElevObstruction:

			select {

			case <-networkChannels.WorldViewUpdate:
				setHallLightOnAllPanels(gs)
				orderManagement.RunHallAssigner(gs)

			case <-doorTimer.C:
				if !elevio.GetObstruction() {
					elevio.SetDoorOpenLamp(false)
					setElevState(management.ElevIdle)
				}

			case obstructed := <-elevChannels.Obstruction:
				if !obstructed {
					doorTimer.Reset(doorOpenDuration)
				}
			
			case btnPress := <-elevChannels.BtnPresses:

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

			}
		}
	}
}

// -------------------------------------------------------------------------------------------
// Lights
// -------------------------------------------------------------------------------------------

func setHallLightOnAllPanels(gs *orderManagement.GlobalState) {
	state := gs.GetCopy()

	for floor := 0; floor < management.NumFloors; floor++ {
		for btn := 0; btn < 2; btn++ {
			elevio.SetButtonLamp(
				elevio.ButtonType(btn),
				floor,
				state.HallRequests[floor][btn],
			)
		}
	}
}

func reachedFloorLightsOff(floor int) {
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
}

// -------------------------------------------------------------------------------------------
// State transitions
// -------------------------------------------------------------------------------------------

func setElevState(state management.State) {
	prev := management.Elev.State
	management.Elev.State = state

	switch state {
	case management.ElevMoving:
		onMovingEntry()
	case management.ElevIdle:
		onIdleEntry()
	case management.ElevObstruction:
		onObstructionEntry()
	}

	fmt.Println("STATE CHANGE:", prev, "->", state)
}

func onMovingEntry() {
	elevio.SetDoorOpenLamp(false)
	setMotorFromDir()
}

func onIdleEntry() {
	setMotorStop()
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
// Motor helpers
// -------------------------------------------------------------------------------------------

func setMotorStop() {
	elevio.SetMotorDirection(elevio.MD_Stop)
}

func setMotorFromDir() {
	switch management.Elev.MoveDir {
	case management.DirUp:
		elevio.SetMotorDirection(elevio.MD_Up)
	case management.DirDown:
		elevio.SetMotorDirection(elevio.MD_Down)
	default:
		elevio.SetMotorDirection(elevio.MD_Stop)
	}
}
