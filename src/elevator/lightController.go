package elevator

import (
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/state"
)

// Turns off all hall and cab lights
func InitLights(numFloors int) {
	for floor := 0; floor < numFloors; floor++ {
		elevIO.SetButtonLamp(elevIO.CabButton, floor, false)

		if floor < numFloors-1 {
			elevIO.SetButtonLamp(elevIO.HallUpButton, floor, false)
		}
		if floor > 0 {
			elevIO.SetButtonLamp(elevIO.HallDownButton, floor, false)
		}
	}
}

// sets the hall light based on global state. For synchronizing hall lights on all elevs
func setHallLight(gs *state.GlobalState) {
	state := gs.GetCopy()

	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < 2; button++ {
			elevIO.SetButtonLamp(
				elevIO.ButtonType(button),
				floor,
				state.HallRequests[floor][button],
			)
		}
	}
}

// sets cab and hall lights
func SetAllLights(elevator *management.Elevator, gs *state.GlobalState) {
	globalState := gs.GetCopy()

	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < management.NumButtons; button++ {

			order := elevator.Orders[floor][button]

			// for cab-orders
			if button == 2 {
				elevIO.SetButtonLamp(
					elevIO.ButtonType(button),
					floor,
					order.OrderPlaced,
				)


			} else {
				// for hall-orders
				elevIO.SetButtonLamp(
					elevIO.ButtonType(button),
					floor,
					globalState.HallRequests[floor][button],
				)
			}

		}
	}
}

func setFloorIndicator(floor int) {
	elevIO.SetFloorIndicator(floor)
}
