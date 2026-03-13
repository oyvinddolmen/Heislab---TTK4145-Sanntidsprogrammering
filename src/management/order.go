package management

import (
	"heislab/elevator/elevIO"
)

type Order struct {
	OrderPlaced bool
	Floor       int
	ButtonType  elevIO.ButtonType
}

func CreateOrder(buttonPress elevIO.ButtonEvent) Order {
	order := Order{
		OrderPlaced: true,
		Floor:       buttonPress.Floor,
		ButtonType:  buttonPress.Button,
	}
	return order
}

func (order Order) IsCab() bool {
	return order.ButtonType == elevIO.CabButton
}

func (order Order) IsHallUp() bool {
	return order.ButtonType == elevIO.HallUpButton
}

func (order Order) IsHallDown() bool {
	return order.ButtonType == elevIO.HallDownButton
}

func (order Order) IsActive() bool {
	return order.OrderPlaced
}



