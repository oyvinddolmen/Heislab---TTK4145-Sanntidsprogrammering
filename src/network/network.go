package network

import (
	"heislab/network/bcast"
	"heislab/network/peers"
	"heislab/state"
)

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
	GlobalStateTx chan<- state.GlobalStateData
	GlobalStateRx <-chan state.GlobalStateData

	// World view update
	WorldViewUpdate chan bool
}

// Initializes network goroutines for peer discovery and global state broadcasts
func InitNetwork(config PortConfig) NetworkConn {
	// --- peer discovery channels ---
	heartbeatEnabled := make(chan bool, 1)
	peerUpdates := make(chan peers.PeerUpdate, 16)

	go peers.Transmitter(config.PeerDiscoveryPort, config.LocalID, heartbeatEnabled)
	go peers.Receiver(config.PeerDiscoveryPort, peerUpdates)

	// --- global state channels ---
	globalStateTx := make(chan state.GlobalStateData, 16)
	globalStateRx := make(chan state.GlobalStateData, 16)
	worldViewUpdate := make(chan bool, 1)

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
