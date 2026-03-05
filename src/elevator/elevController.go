package elevator

// ---------------------------------------------------------------------------------------------------------------------
// In charge of physical elevator functions and driving-logic
// ---------------------------------------------------------------------------------------------------------------------

import (
	"heislab/elevio"
	"heislab/management"
	"heislab/orderManagement"
)

// ---------------------------------------------------------------------------------------------------------------------
// Initalize elevator functions
// ---------------------------------------------------------------------------------------------------------------------

func goToGroundFloor(gs *orderManagement.GlobalState) {
	elevio.SetMotorDirection(elevio.MotorDirDown)
	for elevio.GetFloor() != 0 {
	}
	elevio.SetMotorDirection(elevio.MotorDirStop)
	elevio.SetFloorIndicator(0)
	setElevLastFloor(0)
	setElevFloor(0)
	setElevState(gs, management.ElevIdle)
}

func InitElevator(elevID string, adress string, numFloors int) *orderManagement.GlobalState {
	elevio.Init(adress, numFloors) // To run several simulators, each terminal/simulator needs unique adress
	InitLights(numFloors)
	InitFSM(elevID, numFloors)
	gs := orderManagement.NewGlobalState(elevID)
	goToGroundFloor(gs)
	return gs
}

// ---------------------------------------------------------------------------------------------------------------------
// Driving logic
// ---------------------------------------------------------------------------------------------------------------------

func findMovingDirection(destination int, lastFloor int) elevio.MotorDirection {

	// safety measure
	if destination < 0 {
		return elevio.MotorDirStop
	}

	switch {
	case destination > lastFloor:
		management.Elev.MoveDir = management.DirUp
		return elevio.MotorDirUp
	case destination < lastFloor:
		management.Elev.MoveDir = management.DirDown
		return elevio.MotorDirDown
	default:
		if elevio.GetFloor() == -1 {
			management.Elev.MoveDir = management.DirDown
			return elevio.MotorDirDown // if between two floors, always go down (maybe better solution later, lastMovingDir variable?)
		}
		management.Elev.MoveDir = management.DirIdle
		return elevio.MotorDirStop
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
	elevio.SetMotorDirection(elevio.MotorDirStop)
	management.Elev.MoveDir = management.DirIdle
}

// Sets motordirection in direction of newOrder and changes Elev.MoveDir.
func driveToDestination(destination int, lastFloor int) {
	moveDir := findMovingDirection(destination, lastFloor)
	elevio.SetMotorDirection(moveDir)
	setMotorFromDir()
}

// Turns on doorOpenLight
func openDoor() {
	elevio.SetDoorOpenLamp(true)
}
