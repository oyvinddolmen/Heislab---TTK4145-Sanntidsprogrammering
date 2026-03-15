package management

import (
	"heislab/elevator/elevIO"
)

type Order struct {
	isActive    bool
	floor       int
	buttonType  elevIO.ButtonType
}

func CreateOrder(buttonPress elevIO.ButtonEvent) Order {
	order := Order{
		isActive:    true,
		floor:       buttonPress.Floor,
		buttonType:  buttonPress.Button,
	}
	return order
}

// TODO: Not Used
func (order Order) IsCab() bool {
	return order.buttonType == elevIO.CabButton
}
//
func (order Order) IsHallUp() bool {
	return order.buttonType == elevIO.HallUpButton
}
//
func (order Order) IsHallDown() bool {
	return order.buttonType == elevIO.HallDownButton
}

// -------------------------------------------------------------------------------------------
// Set and get functions for elevator
// -------------------------------------------------------------------------------------------

func (order *Order) SetActiveStatus(active bool) { order.isActive = active }
func (order *Order) SetButtonType(button elevIO.ButtonType) { order.buttonType = button }
func (order *Order) SetFloor(floor int) { order.floor = floor }

func (order Order) GetActiveStatus() bool { return order.isActive }
func (order Order) GetFloor() int { return order.floor }
func (order Order) GetButtonType() elevIO.ButtonType { return order.buttonType }

