package elevator

// ---------------------------------------------------------------------------------------------------------------------
// In charge of physical elevator functions and driving-logic 
// ---------------------------------------------------------------------------------------------------------------------

// OWN LIGHTCONTROLLER FILE UNDER ELEVATOR?? !!!!!!!!!!!!!!!!!!!! YES

import (
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
	setElevState(management.ElevIdle)
}

func InitElevator(elevID string, adress string, numFloors int) {
	elevio.Init(adress, numFloors) // To run several simulators, each terminal/simulator needs unique adress
	setElevState(management.ElevInit)
	initLights(numFloors)
	InitFSM(elevID, numFloors)
	goToGroundFloor()
}

// ---------------------------------------------------------------------------------------------------------------------
// Driving logic
// ---------------------------------------------------------------------------------------------------------------------

func findMovingDirection(destination int, lastFloor int) elevio.MotorDirection {

	// safety measure
	if destination < 0 {
		return elevio.MD_Stop
	}

	switch {
	case destination > lastFloor:
		management.Elev.MoveDir = management.DirUp
		return elevio.MD_Up
	case destination < lastFloor:
		management.Elev.MoveDir = management.DirDown
		return elevio.MD_Down
	default:
		if elevio.GetFloor() == -1 {
			management.Elev.MoveDir = management.DirDown
			return elevio.MD_Down // if between two floors, always go down (maybe better solution later, lastMovingDir variable?)
		}
		management.Elev.MoveDir = management.DirIdle
		return elevio.MD_Stop
	}
}

// Checks if elevator has reached floor
func reachedDestination(floor int) bool {
	if management.Elev.State == management.ElevMoving && floor == management.Elev.CurrentOrder.Floor {
		return true
	} else {
		return false
	}
}

// ---------------------------------------------------------------------------------------------------------------------
// Elevator hardware related functions
// ---------------------------------------------------------------------------------------------------------------------

func stopElevator() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	management.Elev.MoveDir = management.DirIdle
}

// Sets motordirection in direction of newOrder and changes Elev.MoveDir.
func driveToDestination(destination int, lastFloor int) {
	moveDir := findMovingDirection(destination, lastFloor)
	elevio.SetMotorDirection(moveDir)
	setMoveDir(management.Direction(moveDir))
}

// Turns on doorOpenLight
func openDoor() {
	elevio.SetDoorOpenLamp(true)
}

