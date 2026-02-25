package elevator

// ---------------------------------------------------------------------------------------------------------------------
// In charge of driving, driving-logic and setting lights
// ---------------------------------------------------------------------------------------------------------------------

// OWN LIGHTCONTROLLER FILE UNDER ELEVATOR?? !!!!!!!!!!!!!!!!!!!! YES

import (
	"fmt"
	"heislab/elevio"
	"heislab/management"
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
	setElevState(management.IDLE)
}

func ElevatorInit(elevIP string, adress string, numFloors int) {
	elevio.Init(adress, numFloors) // To run several simulators, each terminal/simulator needs unique adress
	setElevState(management.INIT)
	lightInit(numFloors)
	InitFSM(elevIP, numFloors)
	goToGroundFloor()
}

// ---------------------------------------------------------------------------------------------------------------------
// Initalize lights functions
// ---------------------------------------------------------------------------------------------------------------------

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

func findMovingDirection(destination int, lastFloor int) elevio.MotorDirection {

	// safety measure
	if destination < 0 {
		fmt.Println("Destination -1/noOrder")
		return elevio.MD_Stop
	}

	switch {
	case destination > lastFloor:
		management.Elev.MoveDir = management.Dir_Up
		return elevio.MD_Up
	case destination < lastFloor:
		management.Elev.MoveDir = management.Dir_Down
		return elevio.MD_Down
	default:
		if elevio.GetFloor() == -1 {
			management.Elev.MoveDir = management.Dir_Down
			return elevio.MD_Down // if between two floors, always go down (maybe better solution later, lastMovingDir variable?)
		}
		management.Elev.MoveDir = management.Dir_Idle
		return elevio.MD_Stop
	}
}

// ---------------------------------------------------------------------------------------------------------------------
// Driving functions
// ---------------------------------------------------------------------------------------------------------------------

// check if elevator have reached floor
func reachedDestination(floor int) bool {
	if management.Elev.State == management.MOVING && floor == management.Elev.CurrentOrder.Floor {
		return true
	} else {
		return false
	}
}

// turns off lights when reaching destination floor
func reachedFloorLightsOff(floor int) {
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
}

func stopElevator() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	management.Elev.MoveDir = management.Dir_Idle
}

// sets motordirection in direction of newOrder and changes Elev.MoveDir.
func driveToDestination(destination int, lastFloor int) {
	fmt.Println("Destination:", destination, " ||| lastFloor:", lastFloor)
	moveDir := findMovingDirection(destination, lastFloor)
	elevio.SetMotorDirection(moveDir)
	setMoveDir(management.Direction(moveDir))

	fmt.Println("Current destination", destination)
	fmt.Println("moveDir:", moveDir)
}

// turns on doorOpenLight
func openDoor() {
	elevio.SetDoorOpenLamp(true)
	fmt.Println("Doors opening")
}
