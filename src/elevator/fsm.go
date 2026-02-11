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

var Elev Elevator

// -------------------------------------------------------------------------------------------
// Initialize state-machine
// -------------------------------------------------------------------------------------------

func InitFSM(elevID int, numFloors int) {
	noOrder := ordermanagment.Order{Floor: -1, ButtonType: -1, Status: -1, Finished: false}
	Elev.ID = elevID
	Elev.Floor = -1
	Elev.MoveDir = Dir_Down
	Elev.CurrentOrder = noOrder
	for i := range numFloors {
		for j := range NumButtons {
			Elev.Orders[i][j].Floor = i
			Elev.Orders[i][j].ButtonType = j
			Elev.Orders[i][j].Status = -1
			Elev.Orders[i][j].Finished = false
			Elev.Orders[i][j].Confirm = false
		}
	}

	fmt.Println("elevID: ", Elev.ID)

}
