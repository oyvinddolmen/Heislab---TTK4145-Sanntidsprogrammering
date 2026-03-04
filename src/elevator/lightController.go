package elevator

import (
	"heislab/elevio"
	"heislab/management"
	"heislab/orderManagement"
)

// Turns off lights when reaching destination floor
func ReachedFloorLightsOff(floor int) {
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
}

// Turns off all hall and cab lights
func InitLights(numFloors int) {
	for i := 0; i < numFloors; i++ {
		elevio.SetButtonLamp(elevio.BT_Cab, i, false)

		if i < numFloors-1 {
			elevio.SetButtonLamp(elevio.BT_HallUp, i, false)
		}
		if i > 0 {
			elevio.SetButtonLamp(elevio.BT_HallDown, i, false)
		}
	}
}

// Sets hall lights based on active orders
func SetHallLightOnAllPanels(gs *orderManagement.GlobalState) {
	for floor := 0; floor < management.NumFloors; floor++ {
		for btn := 0; btn < 2; btn++ {
			// Hent hall-request under mutex via GetCopy()
			state := gs.GetCopy()
			on := state.HallRequests[floor][btn]

			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, on)
		}
	}
}