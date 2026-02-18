package elevator

// ---------------------------------------------------------------------------------------------------------------------
// In charge of driving and setting lights
// ---------------------------------------------------------------------------------------------------------------------

import (
	"fmt"
	"heislab/elevio"
	"heislab/managment"
	/*"heislab/orderManagment"*/
)

// ---------------------------------------------------------------------------------------------------------------------
// Datatypes
// ---------------------------------------------------------------------------------------------------------------------

type ElevChannels struct {
	MotorDirection chan int
	LastFloor      chan int
	Obstruction    chan bool
	StopBtn        chan bool
	BtnPresses     chan elevio.ButtonEvent // getting buttonpresses on the physical control box
	NewOrder       chan managment.Order    // getting new orders locally (somebody places an order on your own elevator)
}

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
}

// ---------------------------------------------------------------------------------------------------------------------
// Initalize lights functions
// ---------------------------------------------------------------------------------------------------------------------

// function that sets all lights on the elevator controll panel
func setLights(channels ElevChannels) {

	for {
		select {

		case obstruction := <-channels.Obstruction:
			elevio.SetDoorOpenLamp(obstruction)
			fmt.Println("Obstruction:", obstruction)

		case stopBtn := <-channels.StopBtn:
			elevio.SetStopLamp(stopBtn)
			fmt.Println("Stop btn:", stopBtn)

		case lastFloor := <-channels.LastFloor:
			elevio.SetFloorIndicator(lastFloor)
			fmt.Println("floor", lastFloor)

			// turn off order lights when reaching a floor
			elevio.SetButtonLamp(elevio.BT_Cab, lastFloor, false)
			elevio.SetButtonLamp(elevio.BT_HallUp, lastFloor, false)
			elevio.SetButtonLamp(elevio.BT_HallDown, lastFloor, false)

		case btnPress := <-channels.BtnPresses:
			/*
			if orderManagment.OrderConfirmed(btnPress) {
				elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)
			}
			*/
			elevio.SetButtonLamp(btnPress.Button, btnPress.Floor, true)

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
