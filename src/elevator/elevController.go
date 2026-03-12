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

// Moves elevator safely to ground floor
func GoToGroundFloor(elev *management.Elevator) {
	elevio.SetMotorDirection(elevio.MotorDirDown)
	for elevio.GetFloor() != 0 {
		time.Sleep(10 * time.Millisecond)
	}
	elevio.SetMotorDirection(elevio.MotorDirStop)
	elevio.SetFloorIndicator(0)
	elev.Floor = 0
	elev.LastFloor = 0
	elev.MoveDir = management.DirIdle
	elev.State = management.ElevIdle
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

type ElevChannels struct {
	NewFloor    chan int
	Obstruction chan bool
	StopBtn     chan bool
	BtnPresses  chan elevio.ButtonEvent // Getting buttonpresses on the physical control box
}