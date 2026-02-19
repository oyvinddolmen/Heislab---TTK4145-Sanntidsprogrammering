package main

import (
	"fmt"
	"heislab/elevator"
	"heislab/elevio"
	"heislab/management"
	"heislab/network"

	"os"
	"strconv"
)

// -------------------------------------------------------------------------------------------
// Main
// -------------------------------------------------------------------------------------------

func main() {
	// -------------------------------------------------------------------------------------------
	// Retrieving ID and network ports on startup
	// -------------------------------------------------------------------------------------------

	if len(os.Args) != 2 {
		fmt.Println("Forgot to ID the elevator")
		fmt.Println("Id the elevator by adding an argument: go run main.go <ID>")
		return
	}

	elevID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("ID must be an integer")
		return
	}

	// -------------------------------------------------------------------------------------------
	// Initializing channels
	// -------------------------------------------------------------------------------------------

	elevChannels := management.ElevChannels{
		MotorDirection:  make(chan int),
		LastFloor:       make(chan int),
		Obstruction:     make(chan bool),
		StopBtn:         make(chan bool),
		BtnPresses:      make(chan elevio.ButtonEvent),
		WorldViewUpdate: make(chan bool),
	}

	// -------------------------------------------------------------------------------------------
	// Initialize network
	// -------------------------------------------------------------------------------------------

	portCfg := network.PortConfig{
		PeerDiscoveryPort: 15657, 	// Random ports, must be same for all
		MessageBcastPort: 15658,
		NodeID: "",
	}

	networkConn := network.InitNetwork(portCfg) // Burde kanskje tas inn i RunElevator eller noe ?

	// -------------------------------------------------------------------------------------------
	// Initialise elevator and run go-functions
	// -------------------------------------------------------------------------------------------

	elevator.ElevatorInit(elevID, "localhost:15657", 4) // localhost:15657" for simulatoren

	go elevator.RunElevator(elevChannels)
	select {}
}
