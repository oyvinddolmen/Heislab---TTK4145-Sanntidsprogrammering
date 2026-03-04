package elevator

// ---------------------------------------------------------------------------------------------------------------------
// In charge of setting lights
// ---------------------------------------------------------------------------------------------------------------------

import (
	"heislab/elevio"
	"heislab/management"
	"heislab/orderManagement"
)

// ---------------------------------------------------------------------------------------------------------------------
// Light functions
// ---------------------------------------------------------------------------------------------------------------------

// Turns off lights when reaching destination floor
func reachedFloorLightsOff(floor int) {
	elevio.SetButtonLamp(elevio.CabButton, floor, false)
	elevio.SetButtonLamp(elevio.HallUpButton, floor, false)
	elevio.SetButtonLamp(elevio.HallDownButton, floor, false)
}

// turns off all hall and cab lights
func initLights(numFloors int) {
	for i := range numFloors {
		elevio.SetButtonLamp(elevio.CabButton, i, false)

		if i != numFloors {
			elevio.SetButtonLamp(elevio.HallUpButton, i, false)
		}
		if i != 0 {
			elevio.SetButtonLamp(elevio.HallDownButton, i, false)
		}
	}
}

// sets hall lights based on active orders
func setHallLightOnAllPanels() {
	for floor := 0; floor < management.NumFloors; floor++ {
		for btn := 0; btn < 2; btn++ {

			elevio.SetButtonLamp(
				elevio.ButtonType(btn),
				floor,
				orderManagement.GlobalState.HallRequests[floor][btn],
			)
		}
	}
}
