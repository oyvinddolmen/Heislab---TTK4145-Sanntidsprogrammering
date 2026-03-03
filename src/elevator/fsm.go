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

func InitFSM(localIP string, NumFloors int) {
	noOrder := management.Order{Floor: -1, ButtonType: -1, ElevIP: "", Finished: false}
	setElevState(management.INIT)
	management.Elev.IP = localIP
	management.Elev.Floor = -1
	management.Elev.LastFloor = 0
	management.Elev.MoveDir = management.Dir_Down
	management.Elev.CurrentOrder = noOrder
	for floor := 0; floor < NumFloors; floor++ {
		for button := 0; button < management.NumButtons; button++ {
			management.Elev.Orders[floor][button].Floor = floor
			management.Elev.Orders[floor][button].ButtonType = elevio.ButtonType(button)
			management.Elev.Orders[floor][button].ElevIP = ""
			management.Elev.Orders[floor][button].Finished = false
			management.Elev.Orders[floor][button].OrderPlaced = false
		}
	}
}

// -------------------------------------------------------------------------------------------
// Timer for door
// -------------------------------------------------------------------------------------------
var doorTimer *time.Timer

const doorOpenDuration = 2 * time.Second

// -------------------------------------------------------------------------------------------
// Running elevator and FSM
// -------------------------------------------------------------------------------------------

func RunElevator(elevChannels management.ElevChannels, networkChannels network.NetworkConn) {
	go elevio.PollFloorSensor(elevChannels.LastFloor)
	go elevio.PollButtons(elevChannels.BtnPresses)
	go elevio.PollStopButton(elevChannels.StopBtn)
	go elevio.PollObstructionSwitch(elevChannels.Obstruction)
	go runFSM(elevChannels, networkChannels)
}

// -------------------------------------------------------------------------------------------
// Running FSM function
// -------------------------------------------------------------------------------------------

func runFSM(elevChannels management.ElevChannels, networkChannels network.NetworkConn) {
	for {
		switch management.Elev.State {

		case management.IDLE:
			select {

			case <-elevChannels.WorldViewUpdate:
				orderManagement.RunHallAssigner()
				driveToDestination(
					management.Elev.CurrentOrder.Floor,
					management.Elev.LastFloor)

				if getMoveDir() != management.Dir_Idle {
					setElevState(management.MOVING)
				}

			case <-elevChannels.Obstruction:
				setElevState(management.OBSTRUCTION)

			case <-elevChannels.StopBtn:
				if management.Elev.Floor != -1 {
					setElevState(management.OBSTRUCTION)
				} else {
					setElevState(management.STOP)
				}

			case btnPress := <-elevChannels.BtnPresses:
				order := orderManagement.CreateOrder(btnPress)

				if order.ButtonType == management.CabButton {
					orderManagement.AddOrderToOrders(order)
					orderManagement.UpdateLocalGlobalState()
				} else {
					orderManagement.AddHallRequestToGlobalState(order)
					orderManagement.IncremtHallRequestVersion(order)
				}

				network.SendGlobalState(networkChannels.GlobalStateTx)
				orderManagement.RunHallAssigner()

				fmt.Println("Valid order floor", order.Floor, "btn:", btnPress.Button)
				elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)

				driveToDestination(
					management.Elev.CurrentOrder.Floor,
					management.Elev.LastFloor,
				)

				if getMoveDir() != management.Dir_Idle {
					setElevState(management.MOVING)
				}
			}

		case management.MOVING:
			select {

			case <-elevChannels.WorldViewUpdate:
				orderManagement.RunHallAssigner()
				driveToDestination(management.Elev.CurrentOrder.Floor, management.Elev.LastFloor)
				if management.Elev.MoveDir != management.Dir_Idle {
					setElevState(management.MOVING)
				}

			case <-elevChannels.StopBtn:
				setElevState(management.STOP)

			case floor := <-elevChannels.LastFloor:
				setElevLastFloor(floor)
				elevio.SetFloorIndicator(floor)

				if reachedDestination(floor) {
					stopElevator()
					orderManagement.CompleteCurrentOrder()
					reachedFloorLightsOff(floor)
					setElevFloor(floor)
					setElevState(management.OBSTRUCTION)
				}

			case btnPress := <-elevChannels.BtnPresses:
				order := orderManagement.CreateOrder(btnPress)

				if order.ButtonType == management.CabButton {
					orderManagement.AddOrderToOrders(order)
					orderManagement.UpdateLocalGlobalState()
				} else {
					orderManagement.AddHallRequestToGlobalState(order)
					orderManagement.IncremtHallRequestVersion(order)
				}

				network.SendGlobalState(networkChannels.GlobalStateTx)
				orderManagement.RunHallAssigner()

				fmt.Println("Valid order floor", order.Floor, "btn:", btnPress.Button)
				elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)

				driveToDestination(
					management.Elev.CurrentOrder.Floor,
					management.Elev.LastFloor,
				)
			}

		case management.STOP:
			select {

			case btnPress := <-elevChannels.BtnPresses:
				order := orderManagement.CreateOrder(btnPress)

				if order.ButtonType == management.CabButton {
					orderManagement.AddOrderToOrders(order)
					orderManagement.UpdateLocalGlobalState()
				} else {
					orderManagement.AddHallRequestToGlobalState(order)
					orderManagement.IncremtHallRequestVersion(order)
				}

				network.SendGlobalState(networkChannels.GlobalStateTx)
				orderManagement.RunHallAssigner()

				fmt.Println("Valid order floor", order.Floor, "btn:", btnPress.Button)
				elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)

			case <-elevChannels.StopBtn:
				setElevState(management.IDLE)

			case <-elevChannels.WorldViewUpdate:
				orderManagement.RunHallAssigner()
			}

		case management.OBSTRUCTION:
			select {

			case <-doorTimer.C:
				if elevio.GetObstruction() {
					// stay in state OBSTRUCTION
				} else {
					elevio.SetDoorOpenLamp(false)
					setElevState(management.IDLE)
				}

			case obstructed := <-elevChannels.Obstruction:
				if !obstructed {
					doorTimer.Reset(doorOpenDuration)
				}
			}
		}
	}
}

// -------------------------------------------------------------------------------------------
// FSM set functions
// -------------------------------------------------------------------------------------------

// sets elev state and calls onStateEntry functions // TODO cahnge to 2 functions
func setElevState(state management.State) {
	prev := management.Elev.State
	management.Elev.State = state

	switch state {
	case management.STOP:
		onStopEntry()

	case management.IDLE:
		onIdleEntry()

	case management.MOVING:
		onMovingEntry()

	case management.OBSTRUCTION:
		onObstructionEntry()
	}

	fmt.Println("STATE CHANGE:", prev, "->", state)
}

func setMoveDir(moveDir management.Direction) {
	management.Elev.MoveDir = moveDir
}

func setElevLastFloor(lastFloor int) {
	management.Elev.LastFloor = lastFloor
}

func setElevFloor(floor int) {
	management.Elev.Floor = floor
}

// -------------------------------------------------------------------------------------------
// FSM get functions
// -------------------------------------------------------------------------------------------

func getMoveDir() management.Direction {
	return management.Elev.MoveDir
}

func getFloor() int {
	return management.Elev.Floor
}

// -------------------------------------------------------------------------------------------
// On-State-Entry functions
// -------------------------------------------------------------------------------------------

// turns on stopLamp, sets Dir_Idle and goes to OBSTRUCTION/STOP state depending on elevator position
func onStopEntry() {
	elevio.SetStopLamp(true)
	setMoveDir(management.Dir_Idle)
	elevio.SetMotorDirection(elevio.MD_Stop)
}

// turns off door-open and stop-light when going to state MOVING
func onMovingEntry() {
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	setElevFloor(-1)
}

// open door, turns off stop lamp, calls RunHallAssigner() and driveToDestination()
func onIdleEntry() {
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	setMoveDir(management.Dir_Idle)

	//orderManagement.RunHallAssigner()
	driveToDestination(
		management.Elev.CurrentOrder.Floor,
		management.Elev.LastFloor)

	if management.Elev.MoveDir != management.Dir_Idle {
		setElevState(management.MOVING)
	}
}

// stops motor, sets moveDir, starts new doorTimer
func onObstructionEntry() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	setMoveDir(management.Dir_Idle)
	openDoor()
	if doorTimer != nil {
		doorTimer.Stop()
	}
	doorTimer = time.NewTimer(doorOpenDuration)
}
