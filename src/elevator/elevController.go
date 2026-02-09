package elevator

import (
	"fmt"
	"heislab/elevio"
	"time"
)

type Direction int

const (
	Dir_Down Direction = -1
	Dir_Idle Direction = 0
	Dir_Up   Direction = 1
)

func ElevatorInit(elevID int, adress string, numFloors int) {
	elevio.Init(adress, numFloors) // To run several simulators, change adress

	elevator := Elevator{
		State: ElevState{
			ID:        elevID,
			direction: Dir_Down,
			floor:     0,
			available: false},

		orders: OrderMatrix{
			placeholder: 1},

		lights: LightMatrix{
			placeholder2: 2},
	}

	fmt.Println("ElevID", elevator.State.ID)
	fmt.Println("Going Up")
	elevio.SetMotorDirection(elevio.MD_Down)

	time.Sleep(5 * time.Second)
	elevio.SetMotorDirection(elevio.MD_Stop)
	fmt.Println("Stopper")
}
