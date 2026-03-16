package management

import (
	"fmt"
	"heislab/elevator/elevIO"
)

const (
	hallUpButton   	   = 0
	hallDownButton     = 1
	NumHallButtonTypes = 2
	CabButton          = 2
	NumButtons         = 3
	NumFloors          = 4
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

type Elevator struct {
	state         State
	id            string
	floor         int // -1 if between floors
	lastFloor     int
	moveDir       Direction
	currentOrder  Order
	lastOrder     Order
	canTakeOrders bool
	orders        [NumFloors][NumButtons]Order
}

type ElevChannels struct {
	NewFloorChannel    chan int
	ObstructionChannel chan bool
	ButtonPressChannel chan elevIO.ButtonEvent
}

func InitElevator(elevID string) (Elevator, ElevChannels) {
	noOrder := CreateOrder(elevIO.ButtonEvent{Floor: 1, Button: elevIO.ButtonType(elevIO.CabButton)})
	noOrder.SetActiveStatus(false)

	elev := Elevator{
		state:        ElevInit,
        id:           elevID,
        floor:        0,
        lastFloor:    0,
        moveDir:      DirIdle,
        currentOrder: noOrder,
        lastOrder:    noOrder,
    }

	// Create elev's order matrix
	for floor := 0; floor < NumFloors; floor++ {
		for button := 0; button < NumButtons; button++ {
			tempOrder := CreateOrder(elevIO.ButtonEvent{Floor: floor, Button: elevIO.ButtonType(button)})
			tempOrder.SetActiveStatus(false)
			elev.orders[floor][button] = tempOrder
		}
	}

	elevChannels := ElevChannels{
		NewFloorChannel:      make(chan int),
		ObstructionChannel:   make(chan bool),
		ButtonPressChannel:   make(chan elevIO.ButtonEvent),
	}

	return elev, elevChannels
}


// -------------------------------------------------------------------------------------------
// Set functions for elevator
// -------------------------------------------------------------------------------------------

func (elev *Elevator) SetState(state State) { elev.state = state }
func (elev *Elevator) SetFloor(floor int) { elev.floor = floor }
func (elev *Elevator) SetLastFloor(lastFloor int) { elev.lastFloor = lastFloor }
func (elev *Elevator) SetMoveDir(moveDir Direction) { elev.moveDir = moveDir }
func (elev *Elevator) SetCurrentOrder(currentOrder Order) { elev.currentOrder = currentOrder }
func (elev *Elevator) SetLastOrder(lastOrder Order) { elev.lastOrder = lastOrder }
func (elev *Elevator) SetCanTakeOrders(canTakeOrders bool) { elev.canTakeOrders = canTakeOrders }
func (elev *Elevator) SetOrder(order Order) { elev.orders[order.GetFloor()][int(order.GetButtonType())] = order }

func (elev *Elevator) SetCurrentOrderActiveStatus(active bool) { elev.currentOrder.SetActiveStatus(active) }
func (elev *Elevator) SetLastOrderButtonType(button elevIO.ButtonType) { elev.lastOrder.SetButtonType(button) }
func (elev *Elevator) SetOrderActiveStatus(floor, button int, active bool) { elev.orders[floor][button].SetActiveStatus(active) }


// -------------------------------------------------------------------------------------------
// Get functions for elevator
// -------------------------------------------------------------------------------------------

func (elev *Elevator) GetState() State { return elev.state }
func (elev *Elevator) GetID() string { return elev.id }
func (elev *Elevator) GetFloor() int { return elev.floor }
func (elev *Elevator) GetLastFloor() int { return elev.lastFloor }
func (elev *Elevator) GetMoveDir() Direction { return elev.moveDir }
func (elev *Elevator) GetCurrentOrder() Order { return elev.currentOrder }
func (elev *Elevator) GetLastOrder() Order { return elev.lastOrder }
func (elev *Elevator) GetCanTakeOrders() bool { return elev.canTakeOrders }
func (elev *Elevator) GetOrder(floor, button int) Order { return elev.orders[floor][button] }

func (elev *Elevator) GetCurrentOrderActiveStatus() bool { return elev.currentOrder.GetActiveStatus() }
func (elev *Elevator) GetCurrentOrderFloor() int { return elev.currentOrder.GetFloor() }
func (elev *Elevator) GetOrderActiveStatus(floor, button int) bool { return elev.orders[floor][button].GetActiveStatus() }


// Printer alle relevante ordre- og tilstandsinformasjon for heisen
// TODO: remove before handing in code
func (elev *Elevator) PrintOrdersDebug() {
	fmt.Println("Elev floor:", elev.floor)
	fmt.Println("MoveDir:", elev.moveDir)
	fmt.Println("LastOrder.ButtonType (0: HallUp , 1: HallDown , 2: Caborder):", elev.GetLastOrder().GetButtonType())

	for floor := 0; floor < NumFloors; floor++ {
		fmt.Println(
			"floor", floor,
			"cab", elev.orders[floor][elevIO.CabButton].GetActiveStatus(),
			"up", elev.orders[floor][elevIO.HallUpButton].GetActiveStatus(),
			"down", elev.orders[floor][elevIO.HallDownButton].GetActiveStatus(),
		)
	}
}
