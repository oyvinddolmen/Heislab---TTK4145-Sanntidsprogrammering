package elevator

import (
	"heislab/orderManagment"
)

// -------------------------------------------------------------------------------------------
// Struct and variables
// -------------------------------------------------------------------------------------------

const (
	NumFloors  = 3
	NumButtons = 3
)

type State int

const (
	INIT      = 1
	IDLE      = 2
	EXECUTING = 3
)

type Elevator struct {
	State
	ID           int
	Floor        int
	MoveDir      Direction
	CurrentOrder orderManagment.Order
	Orders       [NumFloors][3]orderManagment.Order
}

var Elev Elevator

// -------------------------------------------------------------------------------------------
// Initialize state-machine
// -------------------------------------------------------------------------------------------

func InitFSM(elevID int, numFloors int) {
	noOrder := orderManagment.Order{Floor: -1, ButtonType: -1, Status: -1, Finished: false}
	Elev.State = INIT
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
}
