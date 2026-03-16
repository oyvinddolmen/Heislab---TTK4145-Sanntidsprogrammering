package elevator

import (
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/state"
)

// Turns off all hall and cab lights.
func InitLights() {
	for floor := 0; floor < management.NumFloors; floor++ {
		elevIO.SetButtonLamp(elevIO.CabButton, floor, false)

		if floor < management.NumFloors-1 {
			elevIO.SetButtonLamp(elevIO.HallUpButton, floor, false)
		}
		if floor > 0 {
			elevIO.SetButtonLamp(elevIO.HallDownButton, floor, false)
		}
	}
}

// Sets the hall light based on global state.
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

// Sets cab and hall lights
func SetAllLights(elev *management.Elevator, globalState *state.GlobalState) {
	globalStateCopy := globalState.GetCopy()

	for floor := 0; floor < management.NumFloors; floor++ {
		for button := 0; button < management.NumButtons; button++ {
			order := elev.GetOrder(floor, button)

			if order.IsCabOrder() {
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

// TODO: Unødvendig? elevIO er inkludert der denne brukes
func setFloorIndicator(floor int) {
	elevIO.SetFloorIndicator(floor)
}

func setDoorOpenLampIfAtFloor() {
	if elevIO.GetFloor() != -1 {
		elevIO.SetDoorOpenLamp(true)
	}
}
