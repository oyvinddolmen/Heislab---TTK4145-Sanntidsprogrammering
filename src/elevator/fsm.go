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
			case currentOrder := <-channels.NewOrder:
				// broadcast order
				// verify order is received by other elevs
				// calculate who gets the order
				// if this elevator gets order:
				fmt.Println("New order arrived in channel")
				moveDir := findMovingDirection(currentOrder.Floor, management.Elev.LastFloor, management.Elev.Floor)
				elevio.SetMotorDirection(moveDir)
				management.Elev.State = management.EXECUTING

			case obstruction := <-channels.Obstruction:
				// door open functionality
				elevio.SetDoorOpenLamp(obstruction)
				fmt.Println("state", management.Elev.State)

			case floor := <-channels.LastFloor:
				elevio.SetFloorIndicator(floor)

				// reaching the destination -> stop and turn off lights
				if management.Elev.State == management.EXECUTING && floor == management.Elev.CurrentOrder.Floor {

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
				// somebody places and order
				order := management.Order{
					Floor:      btnPress.Floor,
					ButtonType: int(btnPress.Button),
					Status:     -1,
					Finished:   false,
				}

				fmt.Println("order floor", order.Floor)
				orderManagement.HandleNewOrder(order, channels)

				if orderManagement.OrderConfirmed(btnPress) {
					elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)
				}

				// elevator already at the floor
				if elevio.GetFloor() == btnPress.Floor {
					// openDoor()
					elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, false)
				}
			}

		// -------------------------------------------------------------------------------------------
		// CASE: EXECUTING
		// -------------------------------------------------------------------------------------------

		case management.EXECUTING:
			select {
			case stop := <-channels.StopBtn:
				// stop button functionality
				fmt.Println(stop)

			}

		// -------------------------------------------------------------------------------------------
		// CASE: DOOR OPEN ???
		// -------------------------------------------------------------------------------------------

		case management.DOOROPEN:

		}
	}
}
