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

type PortConfig struct {
	PeerDiscoveryPort int    // Used by peers.Transmitter/Receiver (heartbeats)
	MessageBcastPort  int    // Used by bcast.Transmitter/Receiver (global state)
	LocalIP            string // Local IP
}

type NetworkConn struct {
	LocalIP string

	// Peer discovery
	PeerTxEnabled chan<- bool
	PeerUpdates   <-chan peers.PeerUpdate

	// GlobalState messaging
	GlobalStateTx chan<- orderManagement.GlobalStateType
	GlobalStateRx <-chan orderManagement.GlobalStateType
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
	myIP := cfg.LocalIP
	if myIP == "" {
		ip, err := localip.LocalIP()
		if err != nil {
			myIP = fmt.Sprintf("unknown-%d", os.Getpid())
		} else {
			myIP = fmt.Sprintf("%s-%d", ip, os.Getpid())
		}
	}

	// --- peer discovery channels ---
	peerTxEnabled := make(chan bool, 1)
	peerTxEnabled <- true								// true -> Sends heartbeats
	peerUpdates   := make(chan peers.PeerUpdate, 16)

	go peers.Transmitter(cfg.PeerDiscoveryPort, myIP, peerTxEnabled)
	go peers.Receiver(cfg.PeerDiscoveryPort, peerUpdates)

	// --- global state channels ---
	globalStateTx := make(chan orderManagement.GlobalStateType, 16)
	globalStateRx := make(chan orderManagement.GlobalStateType, 16)

	// bcast wants "interface{} channels", but we hide that here.
	go bcast.Transmitter(cfg.MessageBcastPort, globalStateTx)
	go bcast.Receiver(cfg.MessageBcastPort, globalStateRx)

	return NetworkConn{
		LocalIP:       myIP,
		PeerTxEnabled: peerTxEnabled,
		PeerUpdates:   peerUpdates,
		GlobalStateTx: globalStateTx,
		GlobalStateRx: globalStateRx,
	}
}
