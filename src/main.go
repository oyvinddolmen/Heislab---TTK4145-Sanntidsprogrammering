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

		Hvis du bare vil kjøre én heis:
		start simulator og kjør uten ports.
	*/

	// ------------------------------------------------
	// Flags
	// ------------------------------------------------

	simHost := flag.String("simHost", "localhost", "Simulator host")
	simPort := flag.Int("simPort", 15657, "Simulator port")
	simAddr := flag.String("simAddr", "", "Full simulator address host:port (overrides simHost/simPort)")
	peersPort := flag.Int("peersPort", 15667, "UDP port for peer discovery")
	bcastPort := flag.Int("bcastPort", 15668, "UDP port for global state broadcast")
	elevIDFlag := flag.String("id", "", "Elevator ID (optional)")
	flag.Parse()

	elevAddr := *simAddr
	if elevAddr == "" {
		elevAddr = fmt.Sprintf("%s:%d", *simHost, *simPort)
	}

	elevID := *elevIDFlag

	// ------------------------------------------------
	// Channels
	// ------------------------------------------------

	elevChannels := management.ElevChannels{
		MotorDirection: make(chan int),
		LastFloor:      make(chan int),
		Obstruction:    make(chan bool),
		StopBtn:        make(chan bool),
		BtnPresses:     make(chan elevio.ButtonEvent),
	}

	// ------------------------------------------------
	// Network
	// ------------------------------------------------

	portCfg := network.PortConfig{
		PeerDiscoveryPort: *peersPort,
		MessageBcastPort:  *bcastPort,
		LocalID:           elevID,
	}

	networkConn := network.InitNetwork(portCfg)
	broadcastInterval := 20 * time.Millisecond
	gs := elevator.InitElevator(elevID, elevAddr, management.NumFloors)
	faultTolerance.RecoverOnStartup(gs, networkConn.GlobalStateRx)

	// ------------------------------------------------
	// Network goroutines
	// ------------------------------------------------

	go network.ListenAndMergeGlobalState(
		gs,
		networkConn.GlobalStateRx,
	)

	go network.SendGlobalStatePeriodically(
		gs,
		networkConn.GlobalStateTx,
		broadcastInterval,
	)

	// ------------------------------------------------
	// Start elevator FSM
	// ------------------------------------------------

	go elevator.RunElevator(gs, elevChannels, networkConn)

	select {}
}
