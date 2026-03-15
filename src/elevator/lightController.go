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

// sets the hall light based on global state
func setHallLight(globalState *state.GlobalState) {
	globalStateCopy := globalState.GetCopy()

	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < management.NumHallButtonTypes; button++ {
			elevIO.SetButtonLamp(
				elevIO.ButtonType(button),
				floor,
				globalStateCopy.HallOrders[floor][button],
			)
		}
	}
}

// sets cab and hall lights
func SetAllLights(elevator *management.Elevator, globalState *state.GlobalState) {
	globalStateCopy := globalState.GetCopy()

	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < management.NumButtons; button++ {
			order := elevator.GetOrder(floor, button)

			if button == management.CabButton {
				elevIO.SetButtonLamp(
					elevIO.ButtonType(button),
					floor,
					order.GetActiveStatus(),
				)

			} else {
				elevIO.SetButtonLamp(
					elevIO.ButtonType(button),
					floor,
					globalStateCopy.HallOrders[floor][button],
				)
			}

		}
	}
}

func setFloorIndicator(floor int) {
	elevIO.SetFloorIndicator(floor)
}

func setDoorOpenLampIfNotBetweenFloors() {
	if elevIO.GetFloor() != -1 {
		elevIO.SetDoorOpenLamp(true)
	}
}
