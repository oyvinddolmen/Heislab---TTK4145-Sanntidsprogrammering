package main

import (
	"flag"
	"fmt"
	"heislab/elevator"
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/network"
	"heislab/state"
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
		TODO:

		- vi har to hallUpAndHallDownAndCabAtDifDir variabler, en med stor første bokstav og den andre ikke
		- Kutte power når heisen er mellom etasjer
		- sjeke ut specsa til oppgaven hva heisen skal gjøre dersom obstruksjon går på mens heisen kjører
		- før innlevering: fjerne alle print-og debugfunksjoner. Fjerne README fra utlevert kode
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
	elevChannels := elevator.ElevChannels{
		NewFloor:    make(chan int),
		Obstruction: make(chan bool),
		BtnPresses:  make(chan elevIO.ButtonEvent),
	}

	// --------------------- Init elev and globalState ----------------------
	networkConn := network.InitNetwork(portCfg)
	elevator.InitHardware(elevAddr, management.NumFloors)
	elev := management.InitElevator(elevID, management.NumFloors)
	gs := state.InitGlobalState(&elev, elevID)

	// --------------------- Recover on startup ----------------------
	recoveredElev := network.RecoverOnStartup(&elev, gs, networkConn.GlobalStateRx)
	elevator.GoToNearestFloorUnder(&elev)
	gs.UpdateGlobalState(&elev)
	if recoveredElev {
		elevator.UpdateCurrentOrderAndsafeDrive(&elev, gs)
	}

	// ------------------- Network ---------------------
	broadcastInterval := 20 * time.Millisecond
	go network.ListenAndMergeGlobalState(
		gs,
		networkConn.GlobalStateRx,
		networkConn.WorldViewUpdate,
	)
	go network.StartFailureDetector(gs, networkConn.WorldViewUpdate)
	go network.SendGlobalStatePeriodically(
		&elev,
		gs,
		networkConn.GlobalStateTx,
		broadcastInterval,
	)

	// ----------------- Start elevator FSM -------------------
	go elevator.RunElevator(&elev, gs, elevChannels, networkConn)
	select {}
}
