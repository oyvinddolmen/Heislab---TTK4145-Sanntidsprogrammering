package management

import (
	"fmt"
	"heislab/elevator/elevio"
)

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

func (elev *Elevator) SetMoveDir(moveDir Direction) {
	elev.MoveDir = moveDir
}

func (elev *Elevator) SetElevLastFloor(lastFloor int) {
	elev.LastFloor = lastFloor
}

func (elev *Elevator) SetElevFloor(floor int) {
	elev.Floor = floor
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
func (e *Elevator) PrintOrdersDebug() {
	fmt.Println("Elev floor:", e.Floor)
	fmt.Println("MoveDir:", e.MoveDir)
	fmt.Println("LastOrder.ButtonType (0: HallUp , 1: HallDown , 2: Caborder):", e.LastOrder.ButtonType)

	for f := 0; f < NumFloors; f++ {
		fmt.Println(
			"floor", f,
			"cab", e.Orders[f][elevio.CabButton].OrderPlaced,
			"up", e.Orders[f][elevio.HallUpButton].OrderPlaced,
			"down", e.Orders[f][elevio.HallDownButton].OrderPlaced,
		)
	}
}
