package main

import (
	"flag"
	"fmt"
	"heislab/elevator"
	"heislab/elevator/elevio"
	"heislab/faultTolerance"
	"heislab/management"
	"heislab/network"
	"heislab/orderManagement"
	"time"
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
			TODO

			- Det hender heisen havner i en evig loop og kjører kontinuerlig opp og ned mellom 0. og 3. etasje
				- Dette skjer når den har fått utdelt en order, men en annen heiser tar orderen og fjerner den. Heisen har da fortsatt orderen og kjører dit, men vil aldri stoppe der.

			- Dersom man trykker inne stopp knappen når man har en order, må en annen heis ta over orderen

		 	- Må kunne starte å kjøre igjen etter å ha trykket stopp knappen

			- Dersom man stopper i en hall-button down, men de som går på heisen trykker cab call oppover, skal heisen "si ifra" at
			  den kjører en annen retning (ifølge sepeca til oppgaven) Tobias: Fixed

			- Når en heis starter opp skal den ikke kjøre til etasje 0 og skru av alle lys, men motta hvor den er fra andre heiser.
			  Dersom den ikke mottar noe skal den skru av lys og kjøre til etasje 0 [FIXED]

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
		NewFloor:    make(chan int),
		Obstruction: make(chan bool),
		StopBtn:     make(chan bool),
		BtnPresses:  make(chan elevio.ButtonEvent),
	}

	networkConn := network.InitNetwork(portCfg)

	// ----------- Initializing Elevator State and recovering if possible----------
	elevator.InitElevator(elevID, elevAddr, management.NumFloors)
	gs := orderManagement.InitGlobalState(elevID)
	recoveredElev := faultTolerance.RecoverOnStartup(gs, networkConn.GlobalStateRx)
	elevator.GoToNearestFloorUnder()
	if recoveredElev {
		elevator.UpdateCurrentOrderAndsafeDrive(&management.Elev, gs)
	}

	// ------------------- Network ---------------------
	broadcastInterval := 20 * time.Millisecond
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
