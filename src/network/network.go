package network

import (
	"time"

	"heislab/network/bcast"
	"heislab/network/peers"
	"heislab/orderManagement"
)

// ListenAndMergeGlobalState lytter på innkommende worldViews og oppdaterer globalState
func ListenAndMergeGlobalState(gs *orderManagement.GlobalState, rx <-chan orderManagement.GlobalStateType, worldViewUpdate chan bool) {
	for remoteGlobalState := range rx {
		if gs.NewWorldViev(remoteGlobalState) {
			worldViewUpdate <- true
		}
		
		gs.Merge(remoteGlobalState)
	}
}

// SendGlobalStatePeriodically sender global state med jevne mellomrom
func SendGlobalStatePeriodically(gs *orderManagement.GlobalState, tx chan<- orderManagement.GlobalStateType, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		gs.UpdateLocalGlobalState() // oppdater egen state
		msg := gs.GetCopy()         // ta sikker kopi under mutex
		tx <- msg                   // send
	}
}

// SendGlobalState sender global state én gang
func SendGlobalState(gs *orderManagement.GlobalState, tx chan<- orderManagement.GlobalStateType) {
	gs.UpdateLocalGlobalState()
	msg := gs.GetCopy()
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

	// World view update
	WorldViewUpdate chan bool
}

// InitNetwork initializes network goroutines for peer discovery and global state broadcasts
func InitNetwork(config PortConfig) NetworkConn {
	// --- peer discovery channels ---
	heartbeatEnabled := make(chan bool, 1)
	peerUpdates := make(chan peers.PeerUpdate, 16)

	go peers.Transmitter(config.PeerDiscoveryPort, config.LocalID, heartbeatEnabled)
	go peers.Receiver(config.PeerDiscoveryPort, peerUpdates)

	// --- global state channels ---
	globalStateTx := make(chan orderManagement.GlobalStateType, 16)
	globalStateRx := make(chan orderManagement.GlobalStateType, 16)
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
