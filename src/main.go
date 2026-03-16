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

		RUNNING MULTIPLE ELEVATORS
		go run . -simPort 15657 -peersPort 20001 -bcastPort 20002 -id 1
		go run . -simPort 15667 -peersPort 20001 -bcastPort 20002 -id 2
		go run . -simPort 15677 -peersPort 20001 -bcastPort 20002 -id 3


		På lab: begge programmene skal ha samme simPort 15657
	*/

	/*
		TODO:
		- legg å sende og oppdatere worldView hvert sekund
		- hvert sekund må den kjøre run hallassigner og drive to destination uansett når idle (FIXED IDLE TIMER)
		- vi har to hallUpAndHallDownAndCabAtDifDir variabler, en med stor første bokstav og den andre ikke (FIXED)
		- Kutte power når heisen er mellom etasjer (tror den er good, men må teste på fysisk heis)
		- sjeke ut specsa til oppgaven hva heisen skal gjøre dersom obstruksjon går på mens heisen kjører
		- før innlevering: fjerne alle print-og debugfunksjoner. Fjerne README fra utlevert kode
	*/

	// ---------------- Flag GlobalState for ID and Ports --------------------
	simHost := flag.String("simHost", "localhost", "Simulator host")
	simPort := flag.Int("simPort", 15657, "Simulator port")
	simAddr := flag.String("simAddr", "", "Full simulator address host:port (overrides simHost/simPort)")
	broadcastPort := flag.Int("bcastPort", 15668, "UDP port for global state broadcast")
	elevIDFlag := flag.String("id", "1", "Elevator ID (optional)")
	flag.Parse()

	elevAddr := *simAddr
	if elevAddr == "" {
		elevAddr = fmt.Sprintf("%s:%d", *simHost, *simPort)
	}
	elevID := *elevIDFlag

	// --------------------- Init elev, network and globalState ----------------------
	networkConnection := network.InitNetwork(*broadcastPort)
	elevator.InitHardware(elevAddr)
	elev, elevChannels := management.InitElevator(elevID, management.NumFloors)
	globalState := state.InitGlobalState(&elev, elevID)

	// --------------------- Recover on startup ----------------------
	// TODO: Flytte inn i RunElevator. Må også gå an å få logikken under inn i RecoverOnStartup.
	recoveredElev := network.RecoverOnStartup(&elev, globalState, networkConnection.IncomingGlobalStateChannel)
	elevator.GoToNearestFloorUnder(&elev)
	globalState.UpdateGlobalState(&elev)
	if recoveredElev {
		elevator.UpdateCurrentOrderAndsafeDrive(&elev, globalState)
	}

	// ------------------- Communication ---------------------
	// TODO: Plassere i InitCommunication eller noe ?
	go network.ListenAndMergeGlobalState(
		globalState,
		networkConnection.IncomingGlobalStateChannel,
		networkConnection.WorldViewUpdateChannel,
	)
	go network.StartFailureDetector(globalState, networkConnection.WorldViewUpdateChannel)
	go network.SendGlobalStatePeriodically(
		&elev,
		globalState,
		networkConnection.OutgoingGlobalStateChannel,
	)

	// ----------------- Start elevator FSM -------------------
	go elevator.RunElevator(&elev, globalState, elevChannels, networkConnection)
	select {}
}
