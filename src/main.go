package main

import (
	"flag"
	"fmt"
	"heislab/elevator"
	"heislab/elevio"
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

	// ------------------------------------------------
	// Initialize elevator hardware
	// ------------------------------------------------

	elevator.InitElevator(elevID, elevAddr, management.NumFloors)

	// ------------------------------------------------
	// Initialize GlobalState (THIS is the important line)
	// ------------------------------------------------

	gs := orderManagement.NewGlobalState(elevID)

	// ------------------------------------------------
	// Fault tolerance recovery
	// ------------------------------------------------

	faultTolerance.RecoverOnStartup(gs, networkConn.GlobalStateRx)

	// ------------------------------------------------
	// Network goroutines
	// ------------------------------------------------

	go network.ListenAndMergeGlobalState(
		gs,
		networkConn.GlobalStateRx,
		networkConn.WorldViewUpdate,
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

/*
import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"heislab/hallRequestAssigner"
	"heislab/network"
	"heislab/orderManagement"
)

func main() {
	peersPort := flag.Int("peers", 16657, "UDP port for peer discovery (peers)")
	bcastPort := flag.Int("bcast", 16658, "UDP port for broadcast messages (bcast)")
	flag.Parse()

	netw := network.InitNetwork(network.PortConfig{
		PeerDiscoveryPort: *peersPort,
		MessageBcastPort:  *bcastPort,
		NodeID:            "",
	})

	fmt.Println("=== NETTEST STARTED ===")
	fmt.Println("MyID:", netw.MyID)
	fmt.Printf("Ports: peers=%d bcast=%d\n", *peersPort, *bcastPort)
	fmt.Println("Type a number and press Enter to broadcast an update.")
	fmt.Println("Example: 3")
	fmt.Println("=======================")

	// Print peer updates
	go func() {
		for upd := range netw.PeerUpdates {
			fmt.Printf("\n[PEERS] peers=%v new=%q lost=%v\n", upd.Peers, upd.New, upd.Lost)
			fmt.Print("> ")
		}
	}()

	// Print received GlobalState
	go func() {
		for msg := range netw.GlobalStateRx {
			if msg.SenderID == netw.MyID {
				continue // ignore own broadcasts
			}
			fmt.Printf("\n[RX] GlobalState from %s: %+v\n", msg.SenderID, msg.Payload)
			fmt.Print("> ")
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	counter := 0

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("stdin error:", err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "q" || line == "quit" || line == "exit" {
			fmt.Println("Exiting.")
			return
		}

		// Try parse a floor number (or any int) from input
		floor, err := strconv.Atoi(line)
		if err != nil {
			fmt.Println("Please enter an integer (or type 'q' to quit).")
			continue
		}

		counter++

		// Build a test GlobalStateType.
		//
		// IMPORTANT:
		// This assumes orderManagement.GlobalStateType has the fields you described:
		//   HallRequests [][2]bool
		//   States       map[string]ElevatorStateJSON (or similar)
		//
		// If your actual type differs slightly, just adjust the literal below.
		gs := orderManagement.GlobalStateType{
			HallRequests: make([][2]bool, 4),
			States:       make(map[string]hallRequestAssigner.ElevatorStateJSON),
		}

		// Put something clearly changing into the payload so you can see updates arriving.
		// (Adjust field names if your ElevatorStateJSON differs.)

		gs.States[netw.MyID] = hallRequestAssigner.ElevatorStateJSON{
			Behavior:    "moving",
			Floor:       floor,
			Direction:   "up",
			CabRequests: []bool{true, false, false, false}, // adjust length if needed
		}

		netw.GlobalStateTx <- network.Envelope[orderManagement.GlobalStateType]{
			SenderID: netw.MyID,
			Payload:  gs,
		}

		fmt.Printf("[TX] Broadcasted update #%d (floor=%d)\n", counter, floor)
	}
}
*/
