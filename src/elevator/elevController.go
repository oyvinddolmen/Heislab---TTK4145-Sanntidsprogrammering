package elevator

import (
	"heislab/elevio"
)

func Drive() {
	elevio.SetMotorDirection(elevio.MD_Up)
}
