package elevator

import (
	"heislab/elevator/elevIO"
	"heislab/management"
	"time"
)

// -------------------------------------------------------------------------------------------
// Elevator initialization
// -------------------------------------------------------------------------------------------


func InitHardware(address string, numFloors int) {
    elevIO.InitElevatorIO(address, numFloors)
    InitLights(numFloors)
}

// moves elevator to closest floor under elevators position
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
	elev.Floor = elevIO.GetFloor()
	elev.LastFloor = elevIO.GetFloor()
	elev.MoveDir = management.DirIdle
	elev.State = management.ElevIdle
}

// Sets motor direction based on elevator-struct's MoveDir
func setMotorFromDir(elev *management.Elevator) {
	switch elev.MoveDir {
	case management.DirUp:
		elevIO.SetMotorDirection(elevIO.MotorDirUp)
	case management.DirDown:
		elevIO.SetMotorDirection(elevIO.MotorDirDown)
	default:
		elevIO.SetMotorDirection(elevIO.MotorDirStop)
	}
}

// Sets elevIO motor direction to stop and sets move direction in Elev struct
func stopElevator(elev *management.Elevator) {
	elevIO.SetMotorDirection(elevIO.MotorDirStop)
	elev.SetMoveDir(management.DirIdle)
}

// sets elevIO motordirection to stop
func setMotorStop() {
	elevIO.SetMotorDirection(elevIO.MotorDirStop)
}
