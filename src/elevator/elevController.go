package elevator

import (
	"heislab/elevator/elevIO"
	"heislab/management"
	"time"
)

// -------------------------------------------------------------------------------------------
// Elevator initialization
// -------------------------------------------------------------------------------------------

// Initializes elevator input/output functionality.
func InitHardware(elevAddress string) {
    elevIO.InitElevatorIO(elevAddress, management.NumFloors)
    InitLights()
}

// Moves elevator to closest floor under elevators position.
func GoToNearestFloorUnder(elev *management.Elevator) {
	floor := elevIO.GetFloor()

	if floor == -1 {
		elevIO.SetMotorDirection(elevIO.MotorDirDown)
	}
	for elevIO.GetFloor() == -1 {
		time.Sleep(10 * time.Millisecond)
	}
	
	elevIO.SetMotorDirection(elevIO.MotorDirStop)
	elevIO.SetFloorIndicator(elevIO.GetFloor())
	elev.SetFloor(elevIO.GetFloor())
	elev.SetLastFloor(elevIO.GetFloor())
	elev.SetMoveDir(management.DirIdle)
	elev.SetState(management.ElevIdle)
}


// -------------------------------------------------------------------------------------------
// Elevator helper functions
// -------------------------------------------------------------------------------------------

// Sets motor direction based on elevator struct's MoveDir.
func setMotorFromDir(elev *management.Elevator) {
	switch elev.GetMoveDir() {
	case management.DirUp:
		elevIO.SetMotorDirection(elevIO.MotorDirUp)
	case management.DirDown:
		elevIO.SetMotorDirection(elevIO.MotorDirDown)
	default:
		elevIO.SetMotorDirection(elevIO.MotorDirStop)
	}
}

// Sets elevIO motor direction to stop and sets move direction in Elev struct.
func stopElevator(elev *management.Elevator) {
	elevIO.SetMotorDirection(elevIO.MotorDirStop)
	elev.SetMoveDir(management.DirIdle)
}

// Sets elevIO motor direction to stop.
func setMotorStop() {
	elevIO.SetMotorDirection(elevIO.MotorDirStop)
}
