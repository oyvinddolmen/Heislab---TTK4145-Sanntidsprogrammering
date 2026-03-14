package management

import (
	"fmt"
	"heislab/elevator/elevIO"
)

type Elevator struct {
	State         State
	ID            string
	Floor         int // -1 if between floors
	LastFloor     int
	MoveDir       Direction
	CurrentOrder  Order
	LastOrder     Order
	CanTakeOrders bool
	Orders        [NumFloors][NumButtons]Order
}

type ElevChannels struct {
	NewFloor      chan int
	Obstruction   chan bool
	ButtonPresses chan elevIO.ButtonEvent
}

func InitElevator(elevID string, numFloors int) (Elevator, ElevChannels) {
	noOrder := Order{
		Floor:       1,
		ButtonType:  elevIO.CabButton,
		OrderPlaced: false,
	}

	elev := Elevator{
		ID:           elevID,
		State:        ElevInit,
		Floor:        0,
		LastFloor:    0,
		MoveDir:      DirIdle,
		CurrentOrder: noOrder,
		LastOrder:    noOrder,
	}

	elevChannels := ElevChannels{
		NewFloor:      make(chan int),
		Obstruction:   make(chan bool),
		ButtonPresses: make(chan elevIO.ButtonEvent),
	}

	// create elev's order-matrix
	for floor := 0; floor < numFloors; floor++ {
		for button := 0; button < NumButtons; button++ {
			elev.Orders[floor][button] = Order{
				Floor:       floor,
				ButtonType:  elevIO.ButtonType(button),
				OrderPlaced: false,
			}
		}
	}
	return elev, elevChannels
}

const (
	hallUpButton   = 0
	hallDownButton = 1
	CabButton      = 2
	NumButtons     = 3
	NumFloors      = 4
)

type State int

const (
	ElevInit        = 1
	ElevIdle        = 2
	ElevMoving      = 3
	ElevObstruction = 4
)

type Direction int

const (
	DirDown Direction = -1
	DirIdle Direction = 0
	DirUp   Direction = 1
)

// -------------------------------------------------------------------------------------------
// Get and set functions
// -------------------------------------------------------------------------------------------

func (elev *Elevator) SetMoveDir(moveDir Direction) {
	elev.MoveDir = moveDir
}

func (elev *Elevator) SetElevLastFloor(lastFloor int) {
	elev.LastFloor = lastFloor
}

func (elev *Elevator) SetElevFloor(floor int) {
	elev.Floor = floor
}

func (elev *Elevator) SetElevCanTakeOrders(canTakeOrders bool) {
	elev.CanTakeOrders = canTakeOrders
	fmt.Println("Set elev can take order:", canTakeOrders)
}

func (elev *Elevator) GetFloor() int {
	return elev.Floor
}

func (elev *Elevator) GetElevID() string {
	return elev.ID
}

func (elev *Elevator) AddCabOrderToElevator(order Order) {
	elev.Orders[order.Floor][int(order.ButtonType)] = order
}

func (elev *Elevator) GetCurrentOrderFloor() int {
	return elev.CurrentOrder.Floor
}

// Printer alle relevante ordre- og tilstandsinformasjon for heisen
// TODO: remove before handing in code
func (elev *Elevator) PrintOrdersDebug() {
	fmt.Println("Elev floor:", elev.Floor)
	fmt.Println("MoveDir:", elev.MoveDir)
	fmt.Println("LastOrder.ButtonType (0: HallUp , 1: HallDown , 2: Caborder):", elev.LastOrder.ButtonType)

	for floor := 0; floor < NumFloors; floor++ {
		fmt.Println(
			"floor", floor,
			"cab", elev.Orders[floor][elevIO.CabButton].OrderPlaced,
			"up", elev.Orders[floor][elevIO.HallUpButton].OrderPlaced,
			"down", elev.Orders[floor][elevIO.HallDownButton].OrderPlaced,
		)
	}
}
