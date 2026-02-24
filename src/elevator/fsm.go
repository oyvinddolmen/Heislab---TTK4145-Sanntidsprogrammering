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
				fmt.Println("World view update :)")
				orderManagement.RunHallAssigner()
				driveToDestination(management.Elev.CurrentOrder.Floor, management.Elev.LastFloor)

			case obstruction := <-channels.Obstruction:
				// door open functionality
				elevio.SetDoorOpenLamp(obstruction)
				management.Elev.State = management.DOOROPEN

			case stop := <-channels.StopBtn:
				// stop button functionality
				elevio.SetStopLamp(stop)
				fmt.Println("Stop-btn: ", stop)

			case btnPress := <-channels.BtnPresses:
				// hvis orderen blir mottatt av de andre heisene
				if orderManagement.OrderConfirmed(btnPress) {
					order := orderManagement.CreateOrder(btnPress)
					orderManagement.AddOrderToOrders(order)
					fmt.Println("Valid order floor", order.Floor, "btn: ", btnPress.Button)
					elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)

					orderManagement.RunHallAssigner()
					driveToDestination(management.Elev.CurrentOrder.Floor, management.Elev.LastFloor)
					orderManagement.PrintOrders()
				}

				// elevator already at the floor
				if management.Elev.Floor == btnPress.Floor {
					// openDoor()
					elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, false)
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

			case stop := <-channels.StopBtn:
				// stop button functionality while driving
				fmt.Println(stop)

			case floor := <-channels.LastFloor:
				management.Elev.Floor = floor
				management.Elev.LastFloor = floor
				elevio.SetFloorIndicator(floor)
				fmt.Println("Reached floor:", floor)

				// reaching the destination -> stop, turn off lights and remove order from order-table. State -> IDLE
				if reachedDestination(floor) {
					orderManagement.RemoveOrdersAtFloor(&management.Elev, floor)
					stopElevator()
					reachedFloorLightsOff(floor)
					orderManagement.RunHallAssigner()
					fmt.Println("Ran hall assigner after reaching floor")
					driveToDestination(management.Elev.CurrentOrder.Floor, management.Elev.LastFloor) // NOT SUPPOSED TO BE HERE
					// open doors
					//setElevState(management.IDLE) // SUPPOSED TO BE DOOROPEN [!!!], waiting for implementation of DOOROPEN
					fmt.Println("Reached destination and Switched from MOVING (3) to : ", management.Elev.State)

					// TODO: NYE CURRENTORDER BLIR IKKE ENDRET NÅR MAN NÅR EN ETASJE
				}

			case btnPress := <-channels.BtnPresses:
				// hvis orderen blir mottatt av de andre heisene
				if orderManagement.OrderConfirmed(btnPress) {
					order := orderManagement.CreateOrder(btnPress)
					orderManagement.AddOrderToOrders(order)
					fmt.Println("Valid order floor", order.Floor, "btn: ", btnPress.Button)
					elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)

					orderManagement.RunHallAssigner()
					driveToDestination(management.Elev.CurrentOrder.Floor, management.Elev.LastFloor)
				}

			}

		// -------------------------------------------------------------------------------------------
		// CASE: DOOR OPEN
		// -------------------------------------------------------------------------------------------

		case management.DOOROPEN:
			// when doors closing - driveToDestination()
		}
	}
}

// -------------------------------------------------------------------------------------------
// FSM functions
// -------------------------------------------------------------------------------------------

func setElevState(state management.State) {
	management.Elev.State = state
}
