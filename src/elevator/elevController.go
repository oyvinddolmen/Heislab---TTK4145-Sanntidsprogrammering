package elevator

// ---------------------------------------------------------------------------------------------------------------------
// In charge of controlling the elevator
// ---------------------------------------------------------------------------------------------------------------------

import (
	elevio "Driver-go"
	"fmt"
	"heislab/orderManagment"
)

// ---------------------------------------------------------------------------------------------------------------------
// Datatypes
// ---------------------------------------------------------------------------------------------------------------------

type Direction int

const (
	Dir_Down Direction = -1
	Dir_Idle Direction = 0
	Dir_Up   Direction = 1
)

type ElevChannels struct {
	MotorDirection chan int
	LastFloor      chan int
	Obstruction    chan bool
	StopBtn        chan bool
	LightControl   chan elevio.CabFloorLights // writing to lights for cab- and floor panel
	ButtonPresses  chan elevio.ButtonEvent    // getting buttonpresses on the physical control box
	NewOrder       chan orderManagment.Order  // getting new orders locally (somebody places and order on your own elevator)
}

// ---------------------------------------------------------------------------------------------------------------------
// Initalize elevator and lights
// ---------------------------------------------------------------------------------------------------------------------

func lightInit(numFloors int) {
	for i := range numFloors {
		elevio.SetButtonLamp(elevio.BT_Cab, i, false)
	}
}

func goToGroundFloor() {
	elevio.SetMotorDirection(elevio.MD_Down)
	for elevio.GetFloor() != 0 {
	}
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetFloorIndicator(0)
}

func ElevatorInit(elevID int, adress string, numFloors int) {
	elevio.Init(adress, numFloors) // To run several simulators, each terminal/simulator needs unique adress
	goToGroundFloor()
	lightInit(numFloors)
	InitFSM(elevID, numFloors)
}

func RunElevator(channels ElevChannels) {
	go elevio.PollFloorSensor(channels.LastFloor)
	go elevio.PollButtons(channels.ButtonPresses)
	go elevio.PollStopButton(channels.StopBtn)
	go elevio.PollObstructionSwitch(channels.Obstruction)
	go setLights(channels)

	select {}
}

// function that sets all lights on elevatorbox
func setLights(channels ElevChannels) {
	for obstruction := range channels.Obstruction {
		elevio.SetDoorOpenLamp(obstruction)
		fmt.Println("Obstruction:", obstruction)
	}
}
