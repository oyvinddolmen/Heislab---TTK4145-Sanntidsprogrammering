package elevator

// ---------------------------------------------------------------------------------------------------------------------
// In charge of controlling the elevator
// ---------------------------------------------------------------------------------------------------------------------

import (
	"Driver-go"
)

// ---------------------------------------------------------------------------------------------------------------------
// Datatypes
// ---------------------------------------------------------------------------------------------------------------------

type Direction int

const (
	Dir_Down Direction = -1
	Dir_Idle Direction = 0
	Dir_Up   Direction = 1
)

type ElevChannels struct {
	MotorDirection chan int
	FloorReached   chan int
	// will be more
}

// ---------------------------------------------------------------------------------------------------------------------
// Initalize elevator and lights
// ---------------------------------------------------------------------------------------------------------------------

func lightInit(numFloors int) {
	for i := range numFloors {
		elevio.SetButtonLamp(elevio.BT_Cab, i, false)
	}
}

func goToGroundFloor() {
	elevio.SetMotorDirection(elevio.MD_Down)
	for elevio.GetFloor() != 0 {
	}
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetFloorIndicator(0)
}

func ElevatorInit(elevID int, adress string, numFloors int) {
	elevio.Init(adress, numFloors) // To run several simulators, each terminal/simulator needs unique adress
	goToGroundFloor()
	lightInit(numFloors)
	InitFSM(elevID, numFloors)
}

func RunElevator(channels ElevChannels) {
	
}
