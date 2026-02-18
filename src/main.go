package main

import (
	"fmt"
	"heislab/elevator"
	"heislab/elevio"
	"heislab/network"
	"heislab/management"
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

	elevChannels := elevator.ElevChannels{
		MotorDirection: make(chan int),
		LastFloor:      make(chan int),
		Obstruction:    make(chan bool),
		StopBtn:        make(chan bool),
		BtnPresses:     make(chan elevio.ButtonEvent),
		NewOrder:       make(chan management.Order),
	}

	networkChannels := network.NetworkChannels{
		RcvChannel:   make(chan management.Elevator),
		BcastChannel: make(chan management.Elevator),
	}

	// -------------------------------------------------------------------------------------------
	// Initialise elevator and run go-functions
	// -------------------------------------------------------------------------------------------
	elevator.ElevatorInit(elevID, "localhost:15657", 4) // localhost:15657" for simulatoren

	go elevator.RunElevator(elevChannels)
	go network.RunNetwork(networkChannels)
	select {}
}
