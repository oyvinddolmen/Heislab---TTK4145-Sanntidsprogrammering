package elevator

// ---------------------------------------------------------------------------------------------------------------------
// In charge of driving and setting lights
// ---------------------------------------------------------------------------------------------------------------------

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
	management.Elev.State = management.IDLE
}

func ElevatorInit(elevID int, adress string, numFloors int) {
	elevio.Init(adress, numFloors) // To run several simulators, each terminal/simulator needs unique adress
	lightInit(numFloors)
	goToGroundFloor()
	InitFSM(elevID, numFloors)
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

func findMovingDirection(dest int, lastFloor int, currentFloor int) elevio.MotorDirection {
	floor := currentFloor
	if floor == -1 {
		floor = lastFloor
	}

	switch {
	case dest > currentFloor:
		return elevio.MD_Up

	case dest < currentFloor:
		return elevio.MD_Down

	default:
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
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)

}

func stopElevator() {
	elevio.SetMotorDirection(elevio.MD_Stop)
}

func driveToDestination(destination int, lastFloor int, currentFloor int) {
	moveDir := findMovingDirection(destination, lastFloor, currentFloor)
	elevio.SetMotorDirection(moveDir)
}
