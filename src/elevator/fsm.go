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

const doorOpenDuration = 2 * time.Second

// -------------------------------------------------------------------------------------------
// Initialize FSM
// -------------------------------------------------------------------------------------------

// initializes Elevator struct and creates order matrix
func InitFSM(elevID string, NumFloors int) {
	noOrder := management.Order{Floor: -1, ButtonType: -1, ElevID: ""}

	management.Elev.ID = elevID
	management.Elev.State = management.ElevInit
	management.Elev.Floor = -1 // Between floors
	management.Elev.LastFloor = 0
	management.Elev.MoveDir = management.DirIdle
	management.Elev.CurrentOrder = noOrder

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
	go elevio.PollFloorSensor(elevChannels.LastFloor)
	go elevio.PollButtons(elevChannels.BtnPresses)
	go elevio.PollStopButton(elevChannels.StopBtn)
	go elevio.PollObstructionSwitch(elevChannels.Obstruction)

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
				orderManagement.RunHallAssigner(gs)
				setHallLightOnAllPanels(gs)
				safeDrive()
			case floor := <-elevChannels.LastFloor:
				updateFloor(floor)
			case <-elevChannels.Obstruction:
				setElevState(gs, management.ElevObstruction)
			case <-elevChannels.StopBtn:
				if management.Elev.Floor != -1 {
					setElevState(gs, management.ElevObstruction)
				} else {
					setElevState(gs, management.ElevStop)
				}
			case btn := <-elevChannels.BtnPresses:
				handleButtonPress(gs, btn, networkChannels)
				safeDrive()
			}

		// ----------------- Case: MOVING -------------------------
		case management.ElevMoving:
			select {
			case <-networkChannels.WorldViewUpdate:
				orderManagement.RunHallAssigner(gs)
				setHallLightOnAllPanels(gs)
				safeDrive()
			case floor := <-elevChannels.LastFloor:
				updateFloor(floor)
				if orderManagement.ShouldStop(&management.Elev) {
					stopElevator()
					orderManagement.ClearOrdersAndTurnOfLights(gs)
					gs.Print()
					orderManagement.RunHallAssigner(gs)
					setElevState(gs, management.ElevObstruction)
				}
			case <-elevChannels.StopBtn:
				setElevState(gs, management.ElevStop)
			case btn := <-elevChannels.BtnPresses:
				handleButtonPress(gs, btn, networkChannels)
				safeDrive()
			}

		// ----------------- Case: STOP -------------------------
		case management.ElevStop:
			select {
			case <-networkChannels.WorldViewUpdate:
				orderManagement.RunHallAssigner(gs)
				setHallLightOnAllPanels(gs)
			case btn := <-elevChannels.BtnPresses:
				handleButtonPress(gs, btn, networkChannels)
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
				orderManagement.RunHallAssigner(gs)
				setHallLightOnAllPanels(gs)
			case <-doorTimer.C:
				if !elevio.GetObstruction() {
					elevio.SetDoorOpenLamp(false)
					fmt.Println("Nåværende floor: ", management.Elev.Floor, "og state ", management.Elev.State)
					fmt.Println("Går til idle nå")
					setElevState(gs, management.ElevIdle)
				}
			case btn := <-elevChannels.BtnPresses:
				handleButtonPress(gs, btn, networkChannels)
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

	if order.Floor == management.Elev.Floor {
		setElevState(gs, management.ElevObstruction)
		return
	}

	if order.ButtonType == management.CabButton {
		orderManagement.AddOrderToOrders(order)
		gs.UpdateLocalGlobalState()
	} else {
		gs.AddHallRequest(order)
		gs.IncrementHallRequestVersion(order.Floor, order.ButtonType)
	}

	network.SendGlobalState(gs, networkChannels.GlobalStateTx)
	orderManagement.RunHallAssigner(gs)
	elevio.SetButtonLamp(btn.Button, btn.Floor, true)
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

// updates current order and sets motor-direction
func safeDrive() {
	orderManagement.UpdateCurrentOrder()
	orderManagement.UpdateMoveDir()

	if management.Elev.MoveDir == management.DirIdle {
		setMotorStop()
		return
	}

	setMotorFromDir()
	setElevState(nil, management.ElevMoving)
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
		onIdleEntry()
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
func onIdleEntry() {
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	setMoveDir(management.DirIdle)
	safeDrive()
}

// turns off stop and door open lamp, and sets elevio motor direction
func onMovingEntry() {
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	setMotorFromDir()
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
