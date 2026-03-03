package management

// -------------------------------------------------------------------------------------------
// Struct and variables for Order and Elevator
// -------------------------------------------------------------------------------------------

import (
	"heislab/elevio"
)

// ---------------------------------------------------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------------------------------------------------

const (
	NumFloors  = 4
	NumButtons = 3
	CabButton  = 2
)

type State int

const (
	INIT        = 1
	IDLE        = 2
	MOVING      = 3
	STOP        = 4
	OBSTRUCTION = 5
	OFFLINE     = 6
)

type Direction int

const (
	Dir_Down Direction = -1
	Dir_Idle Direction = 0
	Dir_Up   Direction = 1
)

// ---------------------------------------------------------------------------------------------------------------------
// Structs
// ---------------------------------------------------------------------------------------------------------------------

type Order struct {
	OrderPlaced bool
	Floor       int
	ButtonType  elevio.ButtonType
	ElevIP      string // Empty string if no elevator is assigned, else the ID of the elevator assigned
	Finished    bool
}

type Elevator struct {
	State        State
	IP           string
	Floor        int // -1 if between floors
	LastFloor    int
	MoveDir      Direction
	CurrentOrder Order
	Orders       [NumFloors][NumButtons]Order
}

type ElevChannels struct {
	MotorDirection  chan int
	LastFloor       chan int
	Obstruction     chan bool
	StopBtn         chan bool
	BtnPresses      chan elevio.ButtonEvent // Getting buttonpresses on the physical control box
	WorldViewUpdate chan bool
}

// ---------------------------------------------------------------------------------------------------------------------
// Initiating elevators
// ---------------------------------------------------------------------------------------------------------------------
var Elev Elevator
var OtherElevs []Elevator
