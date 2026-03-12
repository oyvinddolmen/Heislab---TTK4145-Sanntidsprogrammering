package management

import (
	"heislab/elevator/elevio"
)

type Order struct {
	OrderPlaced bool
	Floor       int
	ButtonType  elevio.ButtonType
}

func CreateOrder(btnPress elevio.ButtonEvent) Order {
	order := Order{
		OrderPlaced: true,
		Floor:       btnPress.Floor,
		ButtonType:  btnPress.Button,
	}
	return order
}

func (o Order) IsCab() bool {
	return o.ButtonType == elevio.CabButton
}

func (o Order) IsHallUp() bool {
	return o.ButtonType == elevio.HallUpButton
}

func (o Order) IsHallDown() bool {
	return o.ButtonType == elevio.HallDownButton
}

func (o Order) IsActive() bool {
	return o.OrderPlaced
}



