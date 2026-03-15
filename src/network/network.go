package network

import (
	"heislab/network/broadcast"
	"heislab/state"
)

type NetworkConnection struct {
	OutgoingGlobalStateChannel chan state.GlobalStateData
	IncomingGlobalStateChannel chan state.GlobalStateData
	WorldViewUpdateChannel 	   chan bool
}

// Initializes network channels and goroutines for global state broadcasts
func InitNetwork(broadcastPort int) NetworkConnection {
	networkConnection := NetworkConnection{
		OutgoingGlobalStateChannel: make(chan state.GlobalStateData, 16),
		IncomingGlobalStateChannel: make(chan state.GlobalStateData, 16),
		WorldViewUpdateChannel: 	make(chan bool, 1),
	}
	go broadcast.GlobalStateTransmitter(broadcastPort, networkConnection.OutgoingGlobalStateChannel)
	go broadcast.GlobalStateReceiver(broadcastPort, networkConnection.IncomingGlobalStateChannel)

	return networkConnection
}
