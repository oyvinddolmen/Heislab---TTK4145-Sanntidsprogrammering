package main

import (
	"fmt"
	"os"
	"strconv"
	"heislab/elevator"
)

// -------------------------------------------------------------------------------------------
// Varaiables
// -------------------------------------------------------------------------------------------

// const
// const port = 20000 ?

// -------------------------------------------------------------------------------------------
// Main
// -------------------------------------------------------------------------------------------

func main() {

	// Elevator ID
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

	fmt.Println((elevID))

	// Network channels
		// channels for Receiving and broadcasting

	// FSM
		// initalise finite state machine

	// Elevator
	elevator.ElevatorInit(elevID, "localhost:15657", 4)

}
