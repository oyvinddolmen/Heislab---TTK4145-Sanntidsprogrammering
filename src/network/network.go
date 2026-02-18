package network

// ---------------------------------------------------------------------------------------------------------------------
// Calling of communication functions in network-folder
// ---------------------------------------------------------------------------------------------------------------------

import (
	"fmt"
	"os"
	"time"

	"heislab/management"
	"heislab/orderManagement"
	"heislab/network/bcast"
	"heislab/network/localip"
	"heislab/network/peers"
)

type NetworkChannels struct {
	RcvChannel   chan management.Elevator
	BcastChannel chan management.Elevator
}

func BcastElevInfo(BcastChannel chan management.Elevator) {
	time.Sleep(2 * time.Millisecond)
	BcastChannel <- management.Elev
	// TODO
}

// Envelope wraps any payload with a SenderID so receivers can ignore their own broadcasts.
type Envelope[T any] struct {
	SenderID string
	Payload  T
}

type PortConfig struct {
	PeerDiscoveryPort int // used by peers.Transmitter/Receiver (heartbeats)
	MessageBcastPort  int // used by bcast.Transmitter/Receiver (your actual data)
	NodeID            string
}

type NetworkConn struct {
	MyID string

	// Peer discovery
	PeerTxEnabled chan<- bool
	PeerUpdates   <-chan peers.PeerUpdate

	// GlobalState messaging
	GlobalStateTx chan<- Envelope[orderManagement.GlobalState]
	GlobalStateRx <-chan Envelope[orderManagement.GlobalState]
}


// InitNetwork initializes network goroutines for:
//	1. Peer discovery (Tx and Rx)
//		-> Sends heartbeats and keeps track of peers
//	2. Global state broadcasts (Tx and Rx)
// 
// Also initializes and returns channels for network interactions:
//   - myID: the node ID used on the network
//   - peerTxEnabled: send true/false to enable/disable announcing your presence
//   - peerUpdates: stream of PeerUpdate (New/Lost/Peers)
//	 - globalStateTx: broadcast transmitting channel
// 	 - globalStateRx: broadcast receiving channel

func InitNetwork(cfg PortConfig) NetworkConn {
	myID := cfg.NodeID
	if myID == "" {
		ip, err := localip.LocalIP()
		if err != nil {
			myID = fmt.Sprintf("unknown-%d", os.Getpid())
		} else {
			myID = fmt.Sprintf("%s-%d", ip, os.Getpid())
		}
	}

	// --- peer discovery channels ---
	peerTxEnabled := make(chan bool, 1)
	peerTxEnabled <- true								// true -> Sends heartbeats
	peerUpdates   := make(chan peers.PeerUpdate, 16)

	go peers.Transmitter(cfg.PeerDiscoveryPort, myID, peerTxEnabled)
	go peers.Receiver(cfg.PeerDiscoveryPort, peerUpdates)

	// --- global state channels ---
	globalStateTx := make(chan Envelope[orderManagement.GlobalState], 16)
	globalStateRx := make(chan Envelope[orderManagement.GlobalState], 16)

	// bcast wants "interface{} channels", but we hide that here.
	go bcast.Transmitter(cfg.MessageBcastPort, globalStateTx)
	go bcast.Receiver(cfg.MessageBcastPort, globalStateRx)

	return NetworkConn{
		MyID:          myID,
		PeerTxEnabled: peerTxEnabled,
		PeerUpdates:   peerUpdates,
		GlobalStateTx: globalStateTx,
		GlobalStateRx: globalStateRx,
	}
}
