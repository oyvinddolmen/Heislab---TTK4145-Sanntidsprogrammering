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
	ElevInit    	= 1
	ElevIdle        = 2
	ElevMoving      = 3
	ElevStop        = 4
	ElevObstruction = 5
	ElevOffline     = 6
)

type Direction int

const (
	DirDown Direction = -1
	DirIdle Direction = 0
	DirUp   Direction = 1
)

// ---------------------------------------------------------------------------------------------------------------------
// Structs
// ---------------------------------------------------------------------------------------------------------------------

type Order struct {
	OrderPlaced bool
	Floor       int
	ButtonType  elevio.ButtonType
	ElevID      string // Empty string if no elevator is assigned, else the ID of the elevator assigned
	//Finished    bool
}

type Elevator struct {
	State        State
	ID           string
	Floor        int // -1 if between floors
	LastFloor    int
	MoveDir      Direction
	CurrentOrder Order
	LastOrder    Order
	Orders       [NumFloors][NumButtons]Order
}

type ElevChannels struct {
	MotorDirection chan int
	LastFloor      chan int
	Obstruction    chan bool
	StopBtn        chan bool
	BtnPresses     chan elevio.ButtonEvent // Getting buttonpresses on the physical control box
}

// ---------------------------------------------------------------------------------------------------------------------
// Initiating elevators
// ---------------------------------------------------------------------------------------------------------------------
var Elev Elevator
var OtherElevs []Elevator
