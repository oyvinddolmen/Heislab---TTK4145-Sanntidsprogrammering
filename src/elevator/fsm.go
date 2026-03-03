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

func InitFSM(elevID string, NumFloors int) {
	noOrder := management.Order{Floor: -1, ButtonType: -1, ElevID: "", Finished: false}
	setElevState(management.INIT)
	management.Elev.ID = elevID
	management.Elev.Floor = -1
	management.Elev.LastFloor = 0
	management.Elev.MoveDir = management.Dir_Down
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
				//orderManagement.MergeGlobalState()
				orderManagement.RunHallAssigner()
				driveToDestination(management.Elev.CurrentOrder.Floor, management.Elev.LastFloor)
				if getMoveDir() != management.Dir_Idle {
					setElevState(management.MOVING)
				}

			case <-channels.Obstruction:
				setElevState(management.OBSTRUCTION)

			case <-channels.StopBtn:
				if management.Elev.Floor != -1 {
					setElevState(management.OBSTRUCTION)
				} else {
					setElevState(management.STOP)
				}

			case btnPress := <-channels.BtnPresses:
				
				// elevator already at the floor -> open doors
				if management.Elev.Floor == btnPress.Floor {
					setElevState(management.OBSTRUCTION)
				} else {
					// if order received by other elevators
					//PROBLEM: ORDEREN LEGGES FØRST I STATEN. MÅ FØRST KJØRE HALLASSIGNER I TILFELLE DET ER EN AV DE ANDRE SOM EGNT BØR TA DEN
					//Men HALLASSIGNER RETURNERER BARE ASSIGNED SOM FALSE.. PGA 2 HEISER?
					if orderManagement.OrderConfirmed(btnPress) {
						order := orderManagement.CreateOrder(btnPress)
						orderManagement.AddOrderToOrders(order)
						fmt.Println("Valid order floor", order.Floor, "btn: ", btnPress.Button)
						elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)

						orderManagement.IncremtHallRequestVersion(order)
						orderManagement.UpdateLocalGlobalState()
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
				//orderManagement.MergeGlobalState()
				orderManagement.RunHallAssigner()
				driveToDestination(management.Elev.CurrentOrder.Floor, management.Elev.LastFloor)
				if management.Elev.MoveDir != management.Dir_Idle {
					setElevState(management.MOVING)
				}

			case <-channels.StopBtn:
				setElevState(management.STOP)

			case floor := <-channels.LastFloor:
				setElevLastFloor(floor)
				elevio.SetFloorIndicator(floor)

				// reaching the destination -> stop, turn off lights, remove order. State -> DOOROPEN
				if reachedDestination(floor) {
					stopElevator()
					orderManagement.CompleteCurrentOrder()
					reachedFloorLightsOff(floor)
					setElevFloor(floor)
					setElevState(management.OBSTRUCTION)
				}

			case btnPress := <-channels.BtnPresses:
				// hvis orderen blir mottatt av de andre heisene
				if orderManagement.OrderConfirmed(btnPress) {
					order := orderManagement.CreateOrder(btnPress)
					orderManagement.AddOrderToOrders(order)
					fmt.Println("Valid order floor", order.Floor, "btn: ", btnPress.Button)
					elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)
					
					orderManagement.IncremtHallRequestVersion(order)
					orderManagement.UpdateLocalGlobalState()
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
					orderManagement.IncremtHallRequestVersion(order)
					fmt.Println("Valid order floor", order.Floor, "btn: ", btnPress.Button)
					elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)
				}

			case <-channels.StopBtn:
				setElevState(management.IDLE)

			case <-channels.WorldViewUpdate:
				//orderManagement.MergeGlobalState()
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

// Sets elev state and calls onStateEntry functions // TODO cahnge to 2 functions
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
