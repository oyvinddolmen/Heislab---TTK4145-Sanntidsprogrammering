package orderManagement

// -------------------------------------------------------------------------------------------
// Functions for handling and distributing orders
// -------------------------------------------------------------------------------------------

import (
	"fmt"
	"heislab/elevator/elevio"
	"heislab/management"
)

// -------------------------------------------------------------------------------------------
// Functions managing orders
// -------------------------------------------------------------------------------------------

// checks if any other elevators is attending this order
func OrderNotTaken(order management.Order) bool {
	if order.ElevID == "" {
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
		ElevID:      "",
		//Finished:    false,
	}

	return order
}

func PrintOrders() {
	for f := 0; f < management.NumFloors; f++ {
		for b := 0; b < management.NumButtons; b++ {
			order := management.Elev.Orders[f][b]
			fmt.Printf("Floor: %d Button: %d ID: %s OrderPlaced: %t\n", order.Floor, order.ButtonType, order.ElevID, order.OrderPlaced)
		}
	}
}

func AddCabOrderToElevator(order management.Order) {
	management.Elev.Orders[order.Floor][int(order.ButtonType)] = order
}

func GetCurrentOrderFloor() int {
	return management.Elev.CurrentOrder.Floor
}
