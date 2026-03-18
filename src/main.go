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

	/*
		RUNNING MULTIPLE SIMULATORS
		.\SimElevatorServer.exe --port 15657
		.\SimElevatorServer.exe --port 15667
		.\SimElevatorServer.exe --port 15677

		MAC: ./SimElevatorServer --port 15657
		     go run main.go --id=1 --simPort=15657


		RUNNING MULTIPLE ELEVATORS
		go run . -simPort 15657 -bcastPort 20002 -id 1
		go run . -simPort 15667 -bcastPort 20002 -id 2
		go run . -simPort 15677 -bcastPort 20002 -id 3


		På lab: begge programmene skal ha samme simPort 15657
	*/

	/*
		TODO:
		- Må sende på worldViewUpdate når heisen går offline
		- fjerne case OBSTRUCTION under MOVING
		- fjerne unødvendig ting i inputFromTerminal()
		- Gå gjennom FAT test

		- FØR INNLEVERING:
		- fjerne alle print-og debugfunksjoner.
		- README i src med beskrivelse av koden, P2P og UDP osv..
		- fjerne packet loss, simulator, start_simulators, gitignore, README i simulator,
		  hall request og evt. andre steder
		- Kun kommentarer der man ikke skjønenr hva en funksjon gjør eller annen del av koden
		- fjerne kommentarene i main.go

	*/

	// --------------- Get elevator ID, address and port from terminal input -------------------
	elevID, elevAddress, broadcastPort := inputFromTerminal()

	// --------------------- Initialize elevator, network and globalState ----------------------
	networkChannels := network.InitNetwork(*broadcastPort)
	elevator.InitHardware(elevAddress)
	elev, elevChannels := management.InitElevator(elevID)
	globalState := state.InitGlobalState(&elev, elevID)
	fmt.Println("elevID fra startup: ", globalState.GetLocalID(), elev.GetID())
	// --------------------- Recover on startup ----------------------
	// TODO: Flytte inn i RunElevator. Må også gå an å få logikken under inn i RecoverOnStartup.
	recoveredElev := network.RecoverOnStartup(&elev, globalState, networkChannels.IncomingGlobalStateChannel)
	elevator.GoToNearestFloorUnder(&elev)
	globalState.UpdateGlobalState(&elev)
	if recoveredElev {
		elevator.UpdateCurrentOrderAndDrive(&elev, globalState)
	}
	fmt.Println("elevID etter recovered: ", globalState.GetLocalID(), elev.GetID())
	// ------------------- Communication ---------------------
	network.InitCommunication(&elev, globalState, networkChannels)

	// ----------------- Start elevator FSM -------------------
	go elevator.RunElevator(&elev, globalState, elevChannels, networkChannels)
	select {}
}

// takes inputs as ID, broadcast ports and  from terminal on startup
func inputFromTerminal() (string, string, *int) {
	simPort := flag.Int("simPort", 15657, "Simulator port")
	broadcastPort := flag.Int("bcastPort", 20001, "UDP port for global state broadcast")
	elevIDFlag := flag.String("id", "1", "Elevator ID")
	flag.Parse()

	elevAddress := fmt.Sprintf("%s:%d", "localhost", *simPort)
	elevID := *elevIDFlag

	return elevID, elevAddress, broadcastPort
}
