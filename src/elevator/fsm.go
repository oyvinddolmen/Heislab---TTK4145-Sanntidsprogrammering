package elevator

import (
	"fmt"
	"heislab/elevio"
	"heislab/management"
)

// -------------------------------------------------------------------------------------------
// Struct and variables can be found in managment.go
// -------------------------------------------------------------------------------------------

// -------------------------------------------------------------------------------------------
// Initialize state-machine
// -------------------------------------------------------------------------------------------

func InitFSM(elevID int, NumFloors int) {
	noOrder := management.Order{Floor: -1, ButtonType: -1, Status: -1, Finished: false}
	management.Elev.State = management.INIT
	management.Elev.ID = elevID
	management.Elev.Floor = -1
	management.Elev.MoveDir = management.Dir_Down
	management.Elev.CurrentOrder = noOrder
	for i := 0; i < NumFloors; i++ {
		for j := 0; j < management.NumButtons; j++ {
			management.Elev.Orders[i][j].Floor = i
			management.Elev.Orders[i][j].ButtonType = j
			management.Elev.Orders[i][j].Status = -1
			management.Elev.Orders[i][j].Finished = false
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
		switch management.Elev.State {

		case management.IDLE:
			select {
			case currentOrder := <-channels.NewOrder:
				moveDir := findMovingDirection(currentOrder.Floor, management.Elev.LastFloor, management.Elev.Floor)
				elevio.SetMotorDirection(moveDir)
				management.Elev.State = management.EXECUTING

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

		case management.EXECUTING:
			select {
				case stop := <-channels.StopBtn:
				// stop button functionality 
				fmt.Println(stop)

			}

		}
	}
}
