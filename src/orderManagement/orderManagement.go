package orderManagement

// -------------------------------------------------------------------------------------------
// Functions for handling and distributing orders
// -------------------------------------------------------------------------------------------

import (
	"fmt"
	"heislab/elevio"
	"heislab/management"
)

// -------------------------------------------------------------------------------------------
// Functions managin orders
// -------------------------------------------------------------------------------------------

// function that sends order to other elevators and wait for confirmed from the other elevators
func OrderConfirmed(elevio.ButtonEvent) bool {
	// ....
	// ....

	// hvis eneste levende heis -> return true
	return true
}

// checks if any other elevators is attending this order
func OrderNotTaken(order management.Order) bool {
	if order.ElevIP == "" {
		return true
	} else {
		return false
	}
}

func CreateOrder(btnPress elevio.ButtonEvent) management.Order {
	order := management.Order{
		OrderPlaced: true,
		Floor:       btnPress.Floor,
		ButtonType:  btnPress.Button,
		ElevIP:      "",
		Finished:    false,
	}

	return order
}

func PrintOrders() {
	for f := 0; f < management.NumFloors; f++ {
		for b := 0; b < management.NumButtons; b++ {
			order := management.Elev.Orders[f][b]
			fmt.Printf("Floor: %d Button: %d ID: %d OrderPlaced: %d\n", order.Floor, order.ButtonType, order.ElevIP, order.OrderPlaced)
		}
	}
}

func AddOrderToOrders(order management.Order) {
	management.Elev.Orders[order.Floor][int(order.ButtonType)] = order
}

func RemoveOrdersAtFloor(Elev *management.Elevator, floor int) {
	// TODO: BROADCAST: floor elevator arrived at. ACTION: elevators removes the order from the order-table

	// removing orders from local order-table
	for btn := 0; btn < management.NumButtons; btn++ {

		order := &Elev.Orders[floor][btn]

		if order.OrderPlaced {
			order.OrderPlaced = false
			order.Finished = true
			order.ElevIP = ""
		}
	}
}
