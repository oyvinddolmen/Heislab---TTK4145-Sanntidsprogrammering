package main

import (
	"fmt"
	"heislab/elevator"
	"Network-go/bcast"
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

	elevChannels := elevator.ElevChannels{
		MotorDirection: make(chan int),
		FloorReached:   make(chan int),
		// trenger sikkert flere
	}

	bcast.Transmitter(1, elevChannels.FloorReached)

	// networkChannels := network.NetworkChannels{
	// 	RcvChannel:   make(chan elevator.Elevator),
	// 	BcastChannel: make(chan elevator.Elevator),
	// }

	// -------------------------------------------------------------------------------------------
	// Network
	// -------------------------------------------------------------------------------------------

	// -------------------------------------------------------------------------------------------
	// Initialise elevator and run go-functions
	// -------------------------------------------------------------------------------------------
	elevator.ElevatorInit(elevID, "localhost:15657", 4)

	go elevator.RunElevator(elevChannels)
	// go network.RunNetwork(networkChannels)
	select {}
}
