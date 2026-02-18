package elevator

import (
	"fmt"
	"heislab/elevio"
	"heislab/management"
	"heislab/orderManagement"
)

// -------------------------------------------------------------------------------------------
// Initialize state-machine
// -------------------------------------------------------------------------------------------

func InitFSM(elevID int, NumFloors int) {
	noOrder := management.Order{Floor: -1, ButtonType: -1, Status: -1, Finished: false}
	management.Elev.State = management.INIT
	management.Elev.ID = elevID
	management.Elev.Floor = -1
	management.Elev.LastFloor = -1
	management.Elev.MoveDir = management.Dir_Down
	management.Elev.CurrentOrder = noOrder
	for i := 0; i < NumFloors; i++ {
		for j := 0; j < management.NumButtons; j++ {
			management.Elev.Orders[i][j].Floor = i
			management.Elev.Orders[i][j].ButtonType = j
			management.Elev.Orders[i][j].Status = -1
			management.Elev.Orders[i][j].Finished = false
			// maybe more Order variables need to be filled? Must discuss what to include with group
		}
	}
	management.Elev.State = management.IDLE
}

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

			case <-channels.WorldViewUpdate:
				orderManagement.RunHallAssigner()
				orderManagement.PrintOrders()
				driveToDestination(management.Elev.CurrentOrder.Floor, management.Elev.LastFloor, management.Elev.Floor)
				management.Elev.State = management.MOVING

			case obstruction := <-channels.Obstruction:
				// door open functionality
				elevio.SetDoorOpenLamp(obstruction)
				fmt.Println("State", management.Elev.State)

			case floor := <-channels.LastFloor:
				// reaching the destination -> stop and turn off lights
				elevio.SetFloorIndicator(floor)
				if management.Elev.State == management.MOVING && floor == management.Elev.CurrentOrder.Floor {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
					elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
					elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
					management.Elev.State = management.IDLE
				}

			case stop := <-channels.StopBtn:
				// stop button functionality
				elevio.SetStopLamp(stop)
				fmt.Println("Stop-btn: ", stop)

			case btnPress := <-channels.BtnPresses:
				// hvis orderen blir mottatt av de andre heisene
				if orderManagement.OrderConfirmed(btnPress) {
					order := orderManagement.CreateOrder(btnPress)
					orderManagement.AddOrderToOrders(order)
					fmt.Println("Valid order floor", order.Floor)
					elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)

					// not really supposed to be here, only here for testing
					orderManagement.RunHallAssigner()
					orderManagement.PrintOrders()
					driveToDestination(management.Elev.CurrentOrder.Floor, management.Elev.LastFloor, management.Elev.Floor)
					management.Elev.State = management.MOVING
				}

				// elevator already at the floor
				if elevio.GetFloor() == btnPress.Floor {
					// openDoor()
					elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, false)
				}
			}

		// -------------------------------------------------------------------------------------------
		// CASE: MOVING
		// -------------------------------------------------------------------------------------------

		case management.MOVING:
			select {

			case <-channels.WorldViewUpdate:
				orderManagement.RunHallAssigner()
				orderManagement.PrintOrders()
				driveToDestination(management.Elev.CurrentOrder.Floor, management.Elev.LastFloor, management.Elev.Floor)

			case stop := <-channels.StopBtn:
				// stop button functionality while driving
				fmt.Println(stop)

			case floor := <-channels.LastFloor:
				elevio.SetFloorIndicator(floor)

				// reaching the destination -> stop and turn off lights. State -> IDLE
				if reachedDestination(floor) {
					reachedFloorLightsOff(floor)
					stopElevator()
					management.Elev.State = management.IDLE
					fmt.Println("State: ", management.Elev.State)
				}
			}

		// -------------------------------------------------------------------------------------------
		// CASE: DOOR OPEN ???
		// -------------------------------------------------------------------------------------------

		case management.DOOROPEN:

		}
	}
}
