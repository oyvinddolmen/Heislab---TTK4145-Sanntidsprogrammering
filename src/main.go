package main

import (
	"flag"
	"fmt"
	"heislab/elevator"
	"heislab/management"
	"heislab/network"
	"heislab/state"
)

func main() {

	// Initialize elevator, network and globalState
	elevID, elevAddress, broadcastPort := inputFromTerminal()
	networkChannels := network.InitNetwork(*broadcastPort)
	elevator.InitHardware(elevAddress)
	elev, elevChannels := management.InitElevator(elevID)
	globalState := state.InitGlobalState(&elev, elevID)

	// Check if elev can recover orders on startup
	recoveredElev := network.RecoverOnStartup(&elev, globalState, networkChannels.IncomingGlobalStateChannel)

	// Initialize communication
	network.InitCommunication(&elev, globalState, networkChannels)

	// Start elevator FSM
	elevator.RunElevator(&elev, globalState, elevChannels, networkChannels, recoveredElev)
	select {}
}

// Lets you specify ID, simulator port and communication port in terminal when running the program
func inputFromTerminal() (string, string, *int) {
	simPort := flag.Int("simPort", 15657, "Simulator port")
	broadcastPort := flag.Int("bcastPort", 20001, "UDP port communication")
	elevIDFlag := flag.String("id", "1", "Elevator ID")
	flag.Parse()

	elevAddress := fmt.Sprintf("%s:%d", "localhost", *simPort)
	elevID := *elevIDFlag

	return elevID, elevAddress, broadcastPort
}
