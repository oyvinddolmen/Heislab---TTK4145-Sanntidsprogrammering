package managment

// -------------------------------------------------------------------------------------------
// Struct and variables for Order and Elevator
// -------------------------------------------------------------------------------------------

type Order struct {
	OrderPlaced bool
	Floor       int
	ButtonType  int
	Status      int       // -1 if no elevator is assigned, else the ID of the elevator assigned
	Direction   Direction // Dir_Idle if cab call
	Finished    bool
}

const (
	NumFloors  = 4
	NumButtons = 3
)

type State int

const (
	INIT      = 1
	IDLE      = 2
	EXECUTING = 3
	DOOROPEN  = 4
	OFFLINE   = 5
)

type Direction int

const (
	Dir_Down Direction = -1
	Dir_Idle Direction = 0
	Dir_Up   Direction = 1
)

type Elevator struct {
	State        State
	ID           int
	Floor        int
	LastFloor    int
	MoveDir      Direction
	CurrentOrder Order
	Orders       [NumFloors][NumButtons]Order
}

var Elev Elevator
var OtherElevs []Elevator
