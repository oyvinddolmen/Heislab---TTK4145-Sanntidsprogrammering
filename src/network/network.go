package network

import (
	"heislab/network/broadcast"
	"heislab/state"
)

type PortConfig struct {
	PeerDiscoveryPort int    // Used by peers.Transmitter/Receiver (heartbeats)
	MessageBcastPort  int    // Used by bcast.Transmitter/Receiver (global state)
	LocalID           string
}

type NetworkConnection struct {
	// GlobalState messaging
	OutgoingGlobalStateChannel chan<- state.GlobalStateData
	IncomingGlobalStateChannel <-chan state.GlobalStateData
	WorldViewUpdateChannel 	   chan bool
}

// Initializes network channels and goroutines for global state broadcasts
func InitNetwork(config PortConfig) NetworkConnection {
	outgoingGlobalStateChannel := make(chan state.GlobalStateData, 16)
	incomingGlobalStateChannel := make(chan state.GlobalStateData, 16)
	worldViewUpdateChannel := make(chan bool, 1)

	go broadcast.GlobalStateTransmitter(config.MessageBcastPort, outgoingGlobalStateChannel)
	go broadcast.GlobalStateReceiver(config.MessageBcastPort, incomingGlobalStateChannel)

	return NetworkConnection{
		OutgoingGlobalStateChannel: outgoingGlobalStateChannel,
		IncomingGlobalStateChannel: incomingGlobalStateChannel,
		WorldViewUpdateChannel:  	worldViewUpdateChannel,
	}
}
