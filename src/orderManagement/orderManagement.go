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
// Functions
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
	if order.ElevID == 0 {
		return true
	} else {
		return false
	}
}

func CreateOrder(btnPress elevio.ButtonEvent) management.Order {
	order := management.Order{
		OrderPlaced: true,
		Floor:       btnPress.Floor,
		ButtonType:  int(btnPress.Button),
		ElevID:      -1,
		Finished:    false,
	}

	return order
}

func PrintOrders() {
	for f := 0; f < management.NumFloors; f++ {
		for b := 0; b < management.NumButtons; b++ {
			order := management.Elev.Orders[f][b]
			fmt.Printf("Floor: %d Button: %d ID: %d OrderPlaced: %d\n", order.Floor, order.ButtonType, order.ElevID, order.OrderPlaced)
		}
	}
}

func AddOrderToOrders(order management.Order) {
	management.Elev.Orders[order.Floor][int(order.ButtonType)] = order
}
