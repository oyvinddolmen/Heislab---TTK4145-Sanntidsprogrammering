package elevator

import (
	"fmt"
	"heislab/elevio"
	"heislab/managment"
)

// -------------------------------------------------------------------------------------------
// Struct and variables can be found in managment.go
// -------------------------------------------------------------------------------------------

// -------------------------------------------------------------------------------------------
// Initialize state-machine
// -------------------------------------------------------------------------------------------

func InitFSM(elevID int, NumFloors int) {
	noOrder := managment.Order{Floor: -1, ButtonType: -1, Status: -1, Finished: false}
	managment.Elev.State = managment.INIT
	managment.Elev.ID = elevID
	managment.Elev.Floor = -1
	managment.Elev.MoveDir = managment.Dir_Down
	managment.Elev.CurrentOrder = noOrder
	for i := 0; i < NumFloors; i++ {
		for j := 0; j < managment.NumButtons; j++ {
			managment.Elev.Orders[i][j].Floor = i
			managment.Elev.Orders[i][j].ButtonType = j
			managment.Elev.Orders[i][j].Status = -1
			managment.Elev.Orders[i][j].Finished = false
			// more Order variables need to be filled. Must discuss what to include with group
		}
	}
}

// -------------------------------------------------------------------------------------------
// Running elevator and FSM
// -------------------------------------------------------------------------------------------
func RunElevator(channels ElevChannels) {
	go setLights(channels)
	go elevio.PollFloorSensor(channels.LastFloor)
	go elevio.PollButtons(channels.BtnPresses)
	go elevio.PollStopButton(channels.StopBtn)
	go elevio.PollObstructionSwitch(channels.Obstruction)

	for {
		switch managment.Elev.State {

		case managment.IDLE:
			select {
			case currentOrder := <-channels.NewOrder:
				moveDir := findMovingDirection(currentOrder.Floor, managment.Elev.LastFloor, managment.Elev.Floor)
				elevio.SetMotorDirection(moveDir)
				managment.Elev.State = managment.EXECUTING

			case obstruction := <-channels.Obstruction:
				// door open functionality
				fmt.Println(obstruction)

			case stop := <-channels.StopBtn:
				// stop button functionality 
				fmt.Println(stop)

			case btnPress := <-channels.BtnPresses:
				// somebody pressed a order buttons
				fmt.Println(btnPress)
			}

		case managment.EXECUTING:
			select {
				case stop := <-channels.StopBtn:
				// stop button functionality 
				fmt.Println(stop)

			}

		}
	}
}
