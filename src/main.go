package main

import (
	"fmt"
	"heislab/elevator"
	"heislab/elevio"
	//"heislab/faultTolerance"
	"heislab/management"
	"heislab/network"
	"heislab/orderManagement"

	//"heislab/network"

	"os"
	"strconv"
)

// -------------------------------------------------------------------------------------------
// Main
// -------------------------------------------------------------------------------------------

func main() {
	// -------------------------------------------------------------------------------------------
	// Retrieving ID and network ports on startup
	// -------------------------------------------------------------------------------------------

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

	// -------------------------------------------------------------------------------------------
	// Initializing channels
	// -------------------------------------------------------------------------------------------

	elevChannels := management.ElevChannels{
		MotorDirection:  make(chan int),
		LastFloor:       make(chan int),
		Obstruction:     make(chan bool),
		StopBtn:         make(chan bool),
		BtnPresses:      make(chan elevio.ButtonEvent),
		WorldViewUpdate: make(chan bool),
	}

	// -------------------------------------------------------------------------------------------
	// Initialize network
	// -------------------------------------------------------------------------------------------

	portCfg := network.PortConfig{
		PeerDiscoveryPort: 15657, // Random ports, must be same for all
		MessageBcastPort:  15658,
		NodeID:            "",
	}

	networkConn := network.InitNetwork(portCfg) // Burde kanskje tas inn i RunElevator eller noe ?

	// -------------------------------------------------------------------------------------------
	// Initialise elevator and run go-functions
	// -------------------------------------------------------------------------------------------

	orderManagement.InitGlobalState()
	elevator.ElevatorInit(elevID, "localhost:15657", management.NumFloors) // localhost:15657" for simulatoren

	// TODO ØYVIND: fikse setElevState, blir gjort for mange ganger

	//faultTolerance.RecoverOnStartup(GlobalStateRx)
	
	go elevator.RunElevator(elevChannels)
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
