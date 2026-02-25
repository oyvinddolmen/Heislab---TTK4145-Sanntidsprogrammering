package elevator

import (
	"fmt"
	"heislab/elevio"
	"heislab/management"
	"heislab/orderManagement"
	"time"
)

// -------------------------------------------------------------------------------------------
// Initialize state-machine
// -------------------------------------------------------------------------------------------

func InitFSM(elevID int, NumFloors int) {
	noOrder := management.Order{Floor: -1, ButtonType: -1, ElevID: -1, Finished: false}
	management.Elev.State = management.INIT
	management.Elev.ID = elevID
	management.Elev.Floor = -1
	management.Elev.LastFloor = 0
	management.Elev.MoveDir = management.Dir_Down
	management.Elev.CurrentOrder = noOrder
	for i := 0; i < NumFloors; i++ {
		for j := 0; j < management.NumButtons; j++ {
			management.Elev.Orders[i][j].Floor = i
			management.Elev.Orders[i][j].ButtonType = elevio.ButtonType(j)
			management.Elev.Orders[i][j].ElevID = -1
			management.Elev.Orders[i][j].Finished = false
			management.Elev.Orders[i][j].OrderPlaced = false
		}
	}
	management.Elev.State = management.IDLE
}

// -------------------------------------------------------------------------------------------
// Timer for door
// -------------------------------------------------------------------------------------------
var doorTimer *time.Timer

const doorOpenDuration = 2 * time.Second

// -------------------------------------------------------------------------------------------
// Running elevator and FSM
// -------------------------------------------------------------------------------------------

func RunElevator(channels management.ElevChannels) {
	go elevio.PollFloorSensor(channels.LastFloor)
	go elevio.PollButtons(channels.BtnPresses)
	go elevio.PollStopButton(channels.StopBtn)
	go elevio.PollObstructionSwitch(channels.Obstruction)
	go runFSM(channels)
}

// -------------------------------------------------------------------------------------------
// Running FSM function
// -------------------------------------------------------------------------------------------

func runFSM(channels management.ElevChannels) {
	for {
		switch management.Elev.State {

		// -------------------------------------------------------------------------------------------
		// CASE: IDLE
		// -------------------------------------------------------------------------------------------

		case management.IDLE:
			select {

			// only triggered from outside events (getting broadcast from another elevator)
			case <-channels.WorldViewUpdate:
				orderManagement.RunHallAssigner()
				driveToDestination(management.Elev.CurrentOrder.Floor, management.Elev.LastFloor)
				if management.Elev.MoveDir != management.Dir_Idle {
					setElevState(management.MOVING)
				}

			case obstruction := <-channels.Obstruction:
				fmt.Println("Obstruction: ", obstruction)
				setElevState(management.OBSTRUCTION)

			case stop := <-channels.StopBtn:
				fmt.Println("Stop-btn: ", stop)
				setElevState(management.STOP)

			case btnPress := <-channels.BtnPresses:

				// elevator already at the floor -> open doors
				if management.Elev.Floor == btnPress.Floor {
					setElevState(management.OBSTRUCTION)
				} else {
					// if order received by other elevators
					if orderManagement.OrderConfirmed(btnPress) {
						order := orderManagement.CreateOrder(btnPress)
						orderManagement.AddOrderToOrders(order)
						fmt.Println("Valid order floor", order.Floor, "btn: ", btnPress.Button)
						elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)

						orderManagement.RunHallAssigner()
						driveToDestination(
							management.Elev.CurrentOrder.Floor,
							management.Elev.LastFloor)
						if management.Elev.MoveDir != management.Dir_Idle {
							setElevState(management.MOVING)
						}
					}
				}
			}

		// -------------------------------------------------------------------------------------------
		// CASE: MOVING
		// -------------------------------------------------------------------------------------------

		case management.MOVING:
			select {

			// only triggered from outside events (getting broadcast from another elevator)
			case <-channels.WorldViewUpdate:
				orderManagement.RunHallAssigner()
				driveToDestination(management.Elev.CurrentOrder.Floor, management.Elev.LastFloor)
				if management.Elev.MoveDir != management.Dir_Idle {
					setElevState(management.MOVING)
				}

			case stop := <-channels.StopBtn:
				fmt.Println("Stop-btn", stop)
				setElevState(management.STOP)

			case floor := <-channels.LastFloor:
				management.Elev.Floor = floor
				management.Elev.LastFloor = floor
				elevio.SetFloorIndicator(floor)
				fmt.Println("Reached floor:", floor)

				// reaching the destination -> stop, turn off lights and remove order from order-table. State -> IDLE
				if reachedDestination(floor) {
					stopElevator()
					orderManagement.CompleteCurrentOrder()
					reachedFloorLightsOff(floor)
					setElevState(management.OBSTRUCTION)
				}

			case btnPress := <-channels.BtnPresses:
				// hvis orderen blir mottatt av de andre heisene
				if orderManagement.OrderConfirmed(btnPress) {
					order := orderManagement.CreateOrder(btnPress)
					orderManagement.AddOrderToOrders(order)
					fmt.Println("Valid order floor", order.Floor, "btn: ", btnPress.Button)
					elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)

					orderManagement.RunHallAssigner()

					driveToDestination(
						management.Elev.CurrentOrder.Floor,
						management.Elev.LastFloor)
				}
			}

		// -------------------------------------------------------------------------------------------
		// CASE: STOP BUTTON ACTIVE
		// -------------------------------------------------------------------------------------------

		case management.STOP:
			select {

			case btnPress := <-channels.BtnPresses:

				// if order received by other elevators
				if orderManagement.OrderConfirmed(btnPress) {
					order := orderManagement.CreateOrder(btnPress)
					orderManagement.AddOrderToOrders(order)
					fmt.Println("Valid order floor", order.Floor, "btn: ", btnPress.Button)
					elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)
				}

			case stop := <-channels.StopBtn:
				fmt.Println("Stop-btn: ", stop)
				setElevState(management.IDLE)

			case <-channels.WorldViewUpdate:
				orderManagement.RunHallAssigner()
			}

		// -------------------------------------------------------------------------------------------
		// CASE: OBSTRUCTION/DOOR-OPEN
		// -------------------------------------------------------------------------------------------

		case management.OBSTRUCTION:
			select {

			case <-doorTimer.C:
				if elevio.GetObstruction() {
					// stay in Obstruction state
				} else {
					elevio.SetDoorOpenLamp(false)
					setElevState(management.IDLE)
				}

			case obstructed := <-channels.Obstruction:
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

// sets elev state and calls onStateEntry functions
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

// -------------------------------------------------------------------------------------------
// On-State-Entry functions
// -------------------------------------------------------------------------------------------

// turns on stopLam, sets Dir_Idle and stops elevator
func onStopEntry() {
	elevio.SetStopLamp(true)
	setMoveDir(management.Dir_Idle)
	elevio.SetMotorDirection(elevio.MD_Stop)
	if management.Elev.Floor != -1 {
		openDoor()
		setElevState(management.OBSTRUCTION)
	}
}

// turns off door-open and stop-light when going to state MOVING
func onMovingEntry() {
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
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

func onObstructionEntry() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	setMoveDir(management.Dir_Idle)
	openDoor()
	doorTimer = time.NewTimer(doorOpenDuration)
}
