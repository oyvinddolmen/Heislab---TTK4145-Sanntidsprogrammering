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
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
}

// turns off all hall and cab lights
func initLights(numFloors int) {
	for i := range numFloors {
		elevio.SetButtonLamp(elevio.BT_Cab, i, false)

		if i != numFloors {
			elevio.SetButtonLamp(elevio.BT_HallUp, i, false)
		}
		if i != 0 {
			elevio.SetButtonLamp(elevio.BT_HallDown, i, false)
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
