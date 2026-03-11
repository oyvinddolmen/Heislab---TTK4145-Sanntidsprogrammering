package elevator

import (
	"heislab/elevator/elevio"
	"heislab/management"
	"heislab/orderManagement"
)

// Turns off all hall and cab lights
func InitLights(numFloors int) {
	for floor := 0; floor < numFloors; floor++ {
		elevio.SetButtonLamp(elevio.CabButton, floor, false)

		if floor < numFloors-1 {
			elevio.SetButtonLamp(elevio.HallUpButton, floor, false)
		}
		if floor > 0 {
			elevio.SetButtonLamp(elevio.HallDownButton, floor, false)
		}
	}
}

// sets the hall light based on global state. For synchronizing hall lights on all elevs
func setHallLight(gs *orderManagement.GlobalState) {
	state := gs.GetCopy()

	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < 2; button++ {
			elevio.SetButtonLamp(
				elevio.ButtonType(button),
				floor,
				state.HallRequests[floor][button],
			)
		}
	}
}

// sets cab and hall lights
func SetAllLights(elevator management.Elevator, gs *orderManagement.GlobalState) {
	globalState := gs.GetCopy()

	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < management.NumButtons; button++ {

			order := elevator.Orders[floor][button]

			// for cab-orders
			if button == 2 {
				elevio.SetButtonLamp(
					elevio.ButtonType(button),
					floor,
					order.OrderPlaced,
				)


			} else {
				// for hall-orders
				elevio.SetButtonLamp(
					elevio.ButtonType(button),
					floor,
					globalState.HallRequests[floor][button],
				)
			}

		}
	}
}

func setFloorIndicator(floor int) {
	elevio.SetFloorIndicator(floor)
}
