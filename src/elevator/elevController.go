package elevator

// ---------------------------------------------------------------------------------------------------------------------
// In charge of driving and setting lights
// ---------------------------------------------------------------------------------------------------------------------

import (
	"fmt"
	"heislab/elevio"
	"heislab/managment"
	"heislab/orderManagment"
)

// ---------------------------------------------------------------------------------------------------------------------
// Initalize elevator functions
// ---------------------------------------------------------------------------------------------------------------------

func goToGroundFloor() {
	elevio.SetMotorDirection(elevio.MD_Down)
	for elevio.GetFloor() != 0 {
	}
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetFloorIndicator(0)
	managment.Elev.State = managment.IDLE
}

func ElevatorInit(elevID int, adress string, numFloors int) {
	elevio.Init(adress, numFloors) // To run several simulators, each terminal/simulator needs unique adress
	lightInit(numFloors)
	goToGroundFloor()
	InitFSM(elevID, numFloors)
	managment.Elev.State = managment.IDLE
}

// ---------------------------------------------------------------------------------------------------------------------
// Initalize lights functions
// ---------------------------------------------------------------------------------------------------------------------

// Function that sets all lights on the elevator controll panel
func setLights(channels managment.ElevChannels) {

	for {
		select {

		case obstruction := <-channels.Obstruction:
			elevio.SetDoorOpenLamp(obstruction)
			fmt.Println("Obstruction:", obstruction)
			fmt.Println(managment.Elev.State)

		case stopBtn := <-channels.StopBtn:
			elevio.SetStopLamp(stopBtn)
			fmt.Println("Stop btn:", stopBtn)

		case floor := <-channels.LastFloor:
			elevio.SetFloorIndicator(floor)
			fmt.Println("floor", floor)

			// reaching the destination -> stop and turn off lights
			if managment.Elev.State == managment.EXECUTING && floor == managment.Elev.CurrentOrder.Floor {

				elevio.SetMotorDirection(elevio.MD_Stop)
				elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
				elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
				elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)

				managment.Elev.State = managment.IDLE
			}

		case btnPress := <-channels.BtnPresses:
			if orderManagment.OrderConfirmed(btnPress) {
				elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)
			}

			// elevator already at the floor
			if elevio.GetFloor() == btnPress.Floor {
				// openDoor()
				elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, false)
			}
		}
	}
}

func lightInit(numFloors int) {
	for i := range numFloors {
		elevio.SetButtonLamp(elevio.BT_Cab, i, false)

		if i != numFloors {
			elevio.SetButtonLamp(elevio.BT_HallUp, i, false)
		}
		if i != 0 {
			elevio.SetButtonLamp(elevio.BT_HallDown, i, false)
		}
	}
}

// ---------------------------------------------------------------------------------------------------------------------
// Driving logic
// ---------------------------------------------------------------------------------------------------------------------

func findMovingDirection(dest int, lastFloor int, currentFloor int) elevio.MotorDirection {
	floor := currentFloor
	if floor == -1 {
		floor = lastFloor
	}

	switch {
	case dest > currentFloor:
		return elevio.MD_Up

	case dest < currentFloor:
		return elevio.MD_Down

	default:
		return elevio.MD_Stop
	}
}
