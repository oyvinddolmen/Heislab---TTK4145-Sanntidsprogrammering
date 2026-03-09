package elevator

import (
	"heislab/elevio"
	"heislab/management"
	"time"
)

// -------------------------------------------------------------------------------------------
// Elevator initialization
// -------------------------------------------------------------------------------------------

func InitElevator(elevID string, adress string, numFloors int) {
	elevio.Init(adress, numFloors) 
	InitFSM(elevID, numFloors)
	InitLights(numFloors)
}

// Move elevator safely to ground floor at startup
func GoToGroundFloor() {
	elevio.SetMotorDirection(elevio.MotorDirDown)
	for elevio.GetFloor() != 0 {
		time.Sleep(10 * time.Millisecond)
	}
	elevio.SetMotorDirection(elevio.MotorDirStop)
	elevio.SetFloorIndicator(0)
	management.Elev.Floor = 0
	management.Elev.LastFloor = 0
	management.Elev.MoveDir = management.DirIdle
	management.Elev.State = management.ElevIdle
}

// Sets motor direction based on elevator-struct's MoveDir
func setMotorFromDir() {
	switch management.Elev.MoveDir {
	case management.DirUp:
		elevio.SetMotorDirection(elevio.MotorDirUp)
	case management.DirDown:
		elevio.SetMotorDirection(elevio.MotorDirDown)
	default:
		elevio.SetMotorDirection(elevio.MotorDirStop)
	}
}

// Sets elevio motor direction to stop and sets move direction in Elev struct
func stopElevator() {
	elevio.SetMotorDirection(elevio.MotorDirStop)
	setMoveDir(management.DirIdle)
}

// -------------------------------------------------------------------------------------------
// Utility
// -------------------------------------------------------------------------------------------

// Checks if elevator has reached the current order
//func reachedDestination(floor int) bool {
//	if management.Elev.State == management.ElevMoving && floor == management.Elev.CurrentOrder.Floor {
//		return true
//	}
//	return false
//}
