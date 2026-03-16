package management

import (
	"heislab/elevator/elevIO"
)

type Order struct {
	isActive   bool
	floor      int
	buttonType elevIO.ButtonType
}

func CreateOrder(buttonPress elevIO.ButtonEvent) Order {
	order := Order{
		isActive:    true,
		floor:       buttonPress.Floor,
		buttonType:  buttonPress.Button,
	}
	return order
}

// -------------------------------------------------------------------------------------------
// Set and get functions for elevator
// -------------------------------------------------------------------------------------------

func (order *Order) SetActiveStatus(active bool) { order.isActive = active }
func (order *Order) SetButtonType(button elevIO.ButtonType) { order.buttonType = button }

func (order Order) GetActiveStatus() bool { return order.isActive }
func (order Order) GetFloor() int { return order.floor }
func (order Order) GetButtonType() elevIO.ButtonType { return order.buttonType }


// -------------------------------------------------------------------------------------------
// Order helper functions
// -------------------------------------------------------------------------------------------

func (order Order) IsCabOrder() bool { return order.buttonType == elevIO.CabButton }
func (order Order) IsHallUpOrder() bool { return order.buttonType == elevIO.HallUpButton }
func (order Order) IsHallDownOrder() bool { return order.buttonType == elevIO.HallDownButton }