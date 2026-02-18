package orderManagement

// -------------------------------------------------------------------------------------------
// Functions for handling and distributing orders
// -------------------------------------------------------------------------------------------

import (
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
	return true
}

// checks if any other elevators is attending this order
func OrderNotTaken(order management.Order) bool {
	if order.Status == 0 {
		return true
	} else {
		return false
	}
}

// changes the state currentOrder of given elevator. elevID is the elevator which will get the order
func DistributeOrder(order managment.Order, elevID int, localElevId int) {
	// must change the currentOrder of the correct elevator
}

func HandleNewOrder(order managment.Order, channels managment.ElevChannels) {
	// handle order...
}
