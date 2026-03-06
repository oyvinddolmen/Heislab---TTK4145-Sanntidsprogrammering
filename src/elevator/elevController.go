package elevator

import (
	"heislab/elevio"
	"heislab/management"
	"heislab/orderManagement"
	"time"
)

// -------------------------------------------------------------------------------------------
// Elevator initialization
// -------------------------------------------------------------------------------------------

func InitElevator(elevID string, adress string, numFloors int) *orderManagement.GlobalState {
	elevio.Init(adress, numFloors) // Each simulator/terminal needs a unique address
	InitFSM(elevID, numFloors)
	InitLights(numFloors)
	gs := orderManagement.NewGlobalState(elevID)
	goToGroundFloor(gs)
	return gs
}

// Move elevator safely to ground floor at startup
func goToGroundFloor(gs *orderManagement.GlobalState) {
	elevio.SetMotorDirection(elevio.MotorDirDown)
	for elevio.GetFloor() != 0 {
		time.Sleep(10 * time.Millisecond)
	}
	elevio.SetMotorDirection(elevio.MotorDirStop)
	elevio.SetFloorIndicator(0)
	management.Elev.Floor = 0
	management.Elev.LastFloor = 0
	management.Elev.MoveDir = management.DirIdle
	setElevState(gs, management.ElevIdle)
}

// Sets motor direction based on current MoveDir
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

// Returns true if elevator has reached destination in current order
func reachedDestination(floor int) bool {
	if management.Elev.State == management.ElevMoving && floor == management.Elev.CurrentOrder.Floor {
		return true
	}
	return false
}
