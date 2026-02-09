package elevator

import (
	"fmt"
	ordermanagment "heislab/orderManagment"
)

// -------------------------------------------------------------------------------------------
// Struct and variables
// -------------------------------------------------------------------------------------------

const (
	NumFloors  = 3
	NumButtons = 3
)

type Elevator struct {
	ID           int
	Floor        int
	MoveDir      Direction
	CurrentOrder ordermanagment.Order
	Orders       [NumFloors][3]ordermanagment.Order
}


// -------------------------------------------------------------------------------------------
// Initialize state-machine
// -------------------------------------------------------------------------------------------

func InitFSM(elevID int, numFloors int) {
	noOrder := ordermanagment.Order{Floor: -1, ButtonType: -1, Status: -1, Finished: false}
	elev := Elevator{
		ID:           elevID,
		Floor:        -1,
		MoveDir:      Dir_Down,
		CurrentOrder: noOrder,
	}

	fmt.Println("elevID: ", elev.ID)

}
