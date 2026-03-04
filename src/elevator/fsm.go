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
	noOrder := management.Order{Floor: -1, ButtonType: -1, ElevID: "", Finished: false}
	setElevState(management.ElevInit)
	management.Elev.ID = elevID
	management.Elev.Floor = -1
	management.Elev.LastFloor = 0
	management.Elev.MoveDir = management.DirDown
	management.Elev.CurrentOrder = noOrder

	for floor := 0; floor < NumFloors; floor++ {
		for button := 0; button < management.NumButtons; button++ {
			management.Elev.Orders[floor][button].Floor = floor
			management.Elev.Orders[floor][button].ButtonType = elevio.ButtonType(button)
			management.Elev.Orders[floor][button].ElevID = ""
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

// TODO: dersom man legger inn en cab-order mens dørene er åpne blir den ikke håndtert av heisen
// TODO: heisen blir noen ganger stuck (med døren åpen??) og vil ikke kjøre noe sted

func runFSM(elevChannels management.ElevChannels, networkChannels network.NetworkConn) {
	for {
		switch management.Elev.State {

		case management.ElevIdle:
			select {

			case <-networkChannels.WorldViewUpdate:
				setHallLightOnAllPanels()
				orderManagement.RunHallAssigner()
				driveToDestination(
					management.Elev.CurrentOrder.Floor,
					management.Elev.LastFloor)

				if getMoveDir() != management.DirIdle {
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
				if order.Floor == management.Elev.Floor {
					orderManagement.CompleteCurrentOrder()
				}

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

				if getMoveDir() != management.DirIdle {
					setElevState(management.ElevMoving)
				}

				fmt.Println("Current order: ", management.Elev.CurrentOrder.Floor)
			}

		case management.ElevMoving:
			select {

			case <-networkChannels.WorldViewUpdate:
				setHallLightOnAllPanels()
				orderManagement.RunHallAssigner()
				driveToDestination(management.Elev.CurrentOrder.Floor, management.Elev.LastFloor)
				if management.Elev.MoveDir != management.DirIdle {
					setElevState(management.ElevMoving)
				}

			case <-elevChannels.StopBtn:
				setElevState(management.ElevMoving)

			case floor := <-elevChannels.LastFloor:
				setElevLastFloor(floor)
				elevio.SetFloorIndicator(floor)

				if reachedDestination(floor) {
					stopElevator()
					orderManagement.CompleteCurrentOrder()
					reachedFloorLightsOff(floor)
					setElevFloor(floor)
					setElevState(management.ElevObstruction)
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

				fmt.Println("Current order: ", management.Elev.CurrentOrder.Floor)

			}

		case management.ElevStop:
			select {

			case <-networkChannels.WorldViewUpdate:
				setHallLightOnAllPanels()
				orderManagement.RunHallAssigner()

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
				setElevState(management.ElevIdle)

			}

		case management.ElevObstruction:
			select {

			case <-networkChannels.WorldViewUpdate:
				setHallLightOnAllPanels()
				orderManagement.RunHallAssigner()

			case <-doorTimer.C:
				if elevio.GetObstruction() {
					// stay in state OBSTRUCTION
				} else {
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
					orderManagement.UpdateLocalGlobalState()
				} else {
					orderManagement.AddHallRequestToGlobalState(order)
					orderManagement.IncremtHallRequestVersion(order)
				}

				network.SendGlobalState(networkChannels.GlobalStateTx)
				orderManagement.RunHallAssigner()

				fmt.Println("Valid order floor", order.Floor, "btn:", btnPress.Button)
				elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)

			}
		}
	}
}

// -------------------------------------------------------------------------------------------
// FSM set functions
// -------------------------------------------------------------------------------------------

// Sets elev state and calls onStateEntry functions // TODO cahnge to 2 functions
func setElevState(state management.State) {
	prev := management.Elev.State
	management.Elev.State = state

	switch state {
	case management.ElevStop:
		onStopEntry()

	case management.ElevIdle:
		onIdleEntry()

	case management.ElevMoving:
		onMovingEntry()

	case management.ElevObstruction:
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
	setMoveDir(management.DirIdle)
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
	setMoveDir(management.DirIdle)

	orderManagement.RunHallAssigner()
	driveToDestination(
		management.Elev.CurrentOrder.Floor,
		management.Elev.LastFloor)

	if management.Elev.MoveDir != management.DirIdle {
		setElevState(management.ElevMoving)
	}
}

// stops motor, sets moveDir, starts new doorTimer
func onObstructionEntry() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	setMoveDir(management.DirIdle)
	openDoor()
	if doorTimer != nil {
		doorTimer.Stop()
	}
	doorTimer = time.NewTimer(doorOpenDuration)
}
