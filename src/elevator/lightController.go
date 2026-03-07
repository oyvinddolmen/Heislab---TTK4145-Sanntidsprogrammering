package elevator

import (
	"heislab/elevio"
	"heislab/management"
	"heislab/orderManagement"
)

// Turns off all hall and cab lights
func InitLights(numFloors int) {
	for i := 0; i < numFloors; i++ {
		elevio.SetButtonLamp(elevio.CabButton, i, false)

		if i < numFloors-1 {
			elevio.SetButtonLamp(elevio.HallUpButton, i, false)
		}
		if i > 0 {
			elevio.SetButtonLamp(elevio.HallDownButton, i, false)
		}
	}
}

// sets the hall light based on global state
func setHallLightOnAllPanels(gs *orderManagement.GlobalState) {
	state := gs.GetCopy()

	for floor := 0; floor < management.NumFloors; floor++ {
		for btn := 0; btn < 2; btn++ {
			elevio.SetButtonLamp(
				elevio.ButtonType(btn),
				floor,
				state.HallRequests[floor][btn],
			)
		}
	}
}

func SetAllHallLights(e management.Elevator) {
	for floor := 0; floor < management.NumFloors; floor++ {
		for btn := 0; btn < management.NumButtons; btn++ {

			order := e.Orders[floor][btn]

			elevio.SetButtonLamp(
				elevio.ButtonType(btn),
				floor,
				order.OrderPlaced,
			)
		}
	}
}

func setFloorIndicator(floor int) {
	elevio.SetFloorIndicator(floor)
}
