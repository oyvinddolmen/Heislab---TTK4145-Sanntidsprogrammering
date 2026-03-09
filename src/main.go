package main

import (
	"flag"
	"fmt"
	"heislab/elevator"
	"heislab/elevio"
	"heislab/faultTolerance"
	"heislab/management"
	"heislab/network"
	"time"
)

func main() {

	/*
		RUNNING MULTIPLE SIMULATORS
		.\SimElevatorServer.exe --port 15657
		.\SimElevatorServer.exe --port 15667

		RUNNING MULTIPLE ELEVATORS
		go run . -simPort 15657 -peersPort 20001 -bcastPort 20002 -id 1
		go run . -simPort 15667 -peersPort 20001 -bcastPort 20002 -id 2

		På lab: begge programmene skal ha samme simPort 15657
	*/

	/*
		TODO
		- Når man kjører med flere heiser skal man åpne dør til heis 1 og skru av lys dersom hall button på heis 2 blir presset i etasjen til heis 1
			Delvis fikset, men heislysene blir bare satt på en heis og ikke alltid riktig.

		- Heisen skrur ikke alltid av hall-order lys i riktig retning

		- Dersom man stopper i en hall-button down, men de som går på heisen trykker cab call oppover, skal heisen "si ifra" at
		  den kjører en annen retning (ifølge sepeca til oppgaven)

		- Når en heis starter opp skal den ikke kjøre til etasje 0 og skru av alle lys, men motta hvor den er fra andre heiser.
		  Dersom den ikke mottar noe skal den skru av lys og kjøre til etasje 0

	*/

	// ---------------- Flags for ID and Ports --------------------

	simHost := flag.String("simHost", "localhost", "Simulator host")
	simPort := flag.Int("simPort", 15657, "Simulator port")
	simAddr := flag.String("simAddr", "", "Full simulator address host:port (overrides simHost/simPort)")
	peersPort := flag.Int("peersPort", 15667, "UDP port for peer discovery")
	bcastPort := flag.Int("bcastPort", 15668, "UDP port for global state broadcast")
	elevIDFlag := flag.String("id", "1", "Elevator ID (optional)")
	flag.Parse()

	elevAddr := *simAddr
	if elevAddr == "" {
		elevAddr = fmt.Sprintf("%s:%d", *simHost, *simPort)
	}

	elevID := *elevIDFlag

	// ------------- Port Configuration --------------------
	portCfg := network.PortConfig{
		PeerDiscoveryPort: *peersPort,
		MessageBcastPort:  *bcastPort,
		LocalID:           elevID,
	}

	// --------------------- Channels ----------------------
	elevChannels := management.ElevChannels{
		LastFloor:   make(chan int),
		Obstruction: make(chan bool),
		StopBtn:     make(chan bool),
		BtnPresses:  make(chan elevio.ButtonEvent),
	}

	networkConn := network.InitNetwork(portCfg)

	// ------------------- Network -------------------------
	broadcastInterval := 20 * time.Millisecond
	gs := elevator.InitElevator(elevID, elevAddr, management.NumFloors)
	faultTolerance.RecoverOnStartup(gs, networkConn.GlobalStateRx)
	elevator.UpdateCurrentOrderAndsafeDrive(gs)

	// ------------------- Network goroutines ----------------
	go network.ListenAndMergeGlobalState(
		gs,
		networkConn.GlobalStateRx,
		networkConn.WorldViewUpdate,
	)
	go faultTolerance.StartFailureDetector(gs, networkConn.WorldViewUpdate)
	go network.SendGlobalStatePeriodically(
		gs,
		networkConn.GlobalStateTx,
		broadcastInterval,
	)

	// ----------------- Start elevator FSM -------------------
	go elevator.RunElevator(gs, elevChannels, networkConn)

	select {}
}
