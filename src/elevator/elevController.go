package elevator

import (
	"heislab/elevator/elevio"
	"heislab/management"
	"time"
)

// -------------------------------------------------------------------------------------------
// Elevator initialization
// -------------------------------------------------------------------------------------------


func InitHardware(address string, numFloors int) {
    elevio.Init(address, numFloors)
    InitLights(numFloors)
}

// moves elevator to closest floor under elevators position
func GoToNearestFloorUnder(elev *management.Elevator) {
	floor := elevio.GetFloor()

	if floor == -1 {
		elevio.SetMotorDirection(elevio.MotorDirDown)
	}
	for elevio.GetFloor() == -1 {
		time.Sleep(10 * time.Millisecond)
	}
	elevio.SetMotorDirection(elevio.MotorDirStop)
	elevio.SetFloorIndicator(elevio.GetFloor())
	elev.Floor = elevio.GetFloor()
	elev.LastFloor = elevio.GetFloor()
	elev.MoveDir = management.DirIdle
	elev.State = management.ElevIdle
}

// Sets motor direction based on elevator-struct's MoveDir
func setMotorFromDir(elev *management.Elevator) {
	switch elev.MoveDir {
	case management.DirUp:
		elevio.SetMotorDirection(elevio.MotorDirUp)
	case management.DirDown:
		elevio.SetMotorDirection(elevio.MotorDirDown)
	default:
		elevio.SetMotorDirection(elevio.MotorDirStop)
	}
}

// Sets elevio motor direction to stop and sets move direction in Elev struct
func stopElevator(elev *management.Elevator) {
	elevio.SetMotorDirection(elevio.MotorDirStop)
	elev.SetMoveDir(management.DirIdle)
}

// sets elevio motordirection to stop
func setMotorStop() {
	elevio.SetMotorDirection(elevio.MotorDirStop)
}
