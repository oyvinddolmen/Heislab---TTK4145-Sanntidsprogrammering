package network

// ---------------------------------------------------------------------------------------------------------------------
// Calling of communication functions in network-folder
// ---------------------------------------------------------------------------------------------------------------------

import (
	"time"

	"heislab/network/bcast"
	"heislab/network/peers"
	"heislab/orderManagement"
)

// har ikke testet denne
func ListenAndMergeGlobalState(rx <-chan orderManagement.GlobalStateType, worldViewChannel chan<- struct{}) {
	for msg := range rx {
		orderManagement.MergeGlobalState(msg)

		select {
		case worldViewChannel <- struct{}{}:
		default:
		}
	}
}

// har ikke testet denne
func SendGlobalStatePeriodically(tx chan<- orderManagement.GlobalStateType, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {

		// Oppdater egen state før sending
		orderManagement.UpdateLocalGlobalState()

		// Ta kopi under mutex
		orderManagement.GlobalStateMutex.Lock()
		msg := orderManagement.GlobalState
		orderManagement.GlobalStateMutex.Unlock()

		// Send kopien
		tx <- msg
	}
}

// sends global state once
func SendGlobalState(tx chan<- orderManagement.GlobalStateType) {
	orderManagement.UpdateLocalGlobalState()

	// Ta kopi under mutex
	orderManagement.GlobalStateMutex.Lock()
	msg := orderManagement.GlobalState
	orderManagement.GlobalStateMutex.Unlock()

	// Send kopien
	tx <- msg
}

type PortConfig struct {
	PeerDiscoveryPort int // Used by peers.Transmitter/Receiver (heartbeats)
	MessageBcastPort  int // Used by bcast.Transmitter/Receiver (global state)
	LocalID           string
}

type NetworkConn struct {
	// Peer discovery
	HeartbeatEnabled chan<- bool
	PeerUpdates      <-chan peers.PeerUpdate

	// GlobalState messaging
	GlobalStateTx chan<- orderManagement.GlobalStateType
	GlobalStateRx <-chan orderManagement.GlobalStateType

	// world view update
	WorldViewUpdate chan struct{}
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

func InitNetwork(config PortConfig) NetworkConn {
	// --- peer discovery channels ---
	heartbeatEnabled := make(chan bool, 1)
	peerUpdates := make(chan peers.PeerUpdate, 16)

	go peers.Transmitter(config.PeerDiscoveryPort, config.LocalID, heartbeatEnabled)
	go peers.Receiver(config.PeerDiscoveryPort, peerUpdates)

	// --- global state channels ---
	globalStateTx := make(chan orderManagement.GlobalStateType, 16)
	globalStateRx := make(chan orderManagement.GlobalStateType, 16)
	worldViewUpdate := make(chan struct{}, 1)

	go bcast.Transmitter(config.MessageBcastPort, globalStateTx)
	go bcast.Receiver(config.MessageBcastPort, globalStateRx)

	return NetworkConn{
		HeartbeatEnabled: heartbeatEnabled,
		PeerUpdates:      peerUpdates,
		GlobalStateTx:    globalStateTx,
		GlobalStateRx:    globalStateRx,
		WorldViewUpdate:  worldViewUpdate,
	}
}
