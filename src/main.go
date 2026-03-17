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
		go run . -simPort 15657 -peersPort 20001 -bcastPort 20002 -id 1
		go run . -simPort 15667 -peersPort 20001 -bcastPort 20002 -id 2
		go run . -simPort 15677 -peersPort 20001 -bcastPort 20002 -id 3


		På lab: begge programmene skal ha samme simPort 15657
	*/

	/*
		TODO:
		- obstruction mens den beveger seg i en etasje skal ikke skje, når heisen kjører på floor sensor skal ikke døra åpne seg
		- fjern obstruction under case moving
		- legg å sende og oppdatere worldView hvert sekund
		- hvert sekund må den kjøre run hallassigner og drive to destination uansett når idle (FIXED IDLE TIMER)
		- vi har to hallUpAndHallDownAndCabAtDifDir variabler, en med stor første bokstav og den andre ikke (FIXED)
		- Kutte power når heisen er mellom etasjer (tror den er good, men må teste på fysisk heis)
		- sjeke ut specsa til oppgaven hva heisen skal gjøre dersom obstruksjon går på mens heisen kjører
		- før innlevering: fjerne alle print-og debugfunksjoner. Fjerne README fra utlevert kode
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
		elevator.UpdateCurrentOrderAndsafeDrive(&elev, globalState)
	}
	fmt.Println("elevID etter recovered: ", globalState.GetLocalID(), elev.GetID())
	// ------------------- Communication ---------------------
	network.InitCommunication(&elev, globalState, networkChannels)

	// ----------------- Start elevator FSM -------------------
	go elevator.RunElevator(&elev, globalState, elevChannels, networkChannels)
	select {}
}



// Takes input from terminal on startup and returns elevator ID, network address and state broadcast port.
func inputFromTerminal() (string, string, *int) {
	simHost := flag.String("simHost", "localhost", "Simulator host")
	simPort := flag.Int("simPort", 15657, "Simulator port")
	simAddress := flag.String("simAddr", "", "Full simulator address host:port (overrides simHost/simPort)")
	broadcastPort := flag.Int("bcastPort", 15668, "UDP port for global state broadcast")
	elevIDFlag := flag.String("id", "1", "Elevator ID (optional)")
	flag.Parse()

	elevAddress := *simAddress
	if elevAddress == "" {
		elevAddress = fmt.Sprintf("%s:%d", *simHost, *simPort)
	}
	elevID := *elevIDFlag

	return elevID, elevAddress, broadcastPort
}