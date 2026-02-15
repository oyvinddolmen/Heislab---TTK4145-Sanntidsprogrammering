package elevator

import (
	"heislab/managment"
)

// -------------------------------------------------------------------------------------------
// Struct and variables can be found in managment.go
// -------------------------------------------------------------------------------------------

// -------------------------------------------------------------------------------------------
// Initialize state-machine
// -------------------------------------------------------------------------------------------


func InitFSM(elevID int, NumFloors int) {
	noOrder := managment.Order{Floor: -1, ButtonType: -1, Status: -1, Finished: false}
	managment.Elev.State = managment.INIT
	managment.Elev.ID = elevID
	managment.Elev.Floor = -1
	managment.Elev.MoveDir = managment.Dir_Down
	managment.Elev.CurrentOrder = noOrder
	for i := 0; i < NumFloors; i++ {
		for j := 0; j < managment.NumButtons; j++ {
			managment.Elev.Orders[i][j].Floor = i
			managment.Elev.Orders[i][j].ButtonType = j
			managment.Elev.Orders[i][j].Status = -1
			managment.Elev.Orders[i][j].Finished = false
		}
	}
}
