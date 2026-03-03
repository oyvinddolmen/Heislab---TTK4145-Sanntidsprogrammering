package management

// -------------------------------------------------------------------------------------------
// Struct and variables for Order and Elevator
// -------------------------------------------------------------------------------------------

import (
	"heislab/elevio"
	"os"
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
	Elev_Init        = 1
	Elev_Idle        = 2
	Elev_Moving      = 3
	Elev_Stop        = 4
	Elev_Obstruction = 5
	Elev_Offline     = 6
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
	ElevID      string // Empty string if no elevator is assigned, else the ID of the elevator assigned
	Finished    bool
}

type Elevator struct {
	State        State
	ID           string
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

func GetElevID() string {
	if len(os.Args) < 2 {
		panic("Usage: go run main.go <elevatorID>")
	}

	return os.Args[1]
}
