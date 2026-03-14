package broadcast

import (
	"heislab/network/connection"
	"heislab/state"
	"encoding/json"
	"fmt"
	"net"
)

const bufferSize = 1024

// TODO: Comment
func GlobalStateTransmitter(port int, outgoingGlobalStateChannel <-chan state.GlobalStateData) {
	udpConnection := connection.DialBroadcastUDP(port)
	broadcastAddress, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
	if err != nil {
		panic(fmt.Sprintf("broadcast: failed to resolve broadcast address on port %d: %v", port, err))
	}

	for globalState := range outgoingGlobalStateChannel {
		payload, err := json.Marshal(globalState)
		if err != nil {
			fmt.Printf("broadcast.GlobalStateTransmitter(%d): json.Marshal failed: %v\n", port, err)
			continue
		}

		if len(payload) > bufferSize {
			panic(fmt.Sprintf(
				"Tried to send a message longer than the buffer size (length: %d, buffer size: %d)",
				len(payload), bufferSize,
			))
		}

		_, err = udpConnection.WriteTo(payload, broadcastAddress)
		if err != nil {
			fmt.Printf("broadcast.GlobalStateTransmitter(%d): WriteTo failed: %v\n", port, err)
		}
	}
}

// TODO: Comment
func GlobalStateReceiver(port int, incomingGlobalStateChannel chan<- state.GlobalStateData) {
	var receiveBuffer [bufferSize]byte
	udpConnection := connection.DialBroadcastUDP(port)

	for {
		numBytes, _, err := udpConnection.ReadFrom(receiveBuffer[:])
		if err != nil {
			fmt.Printf("broadcast.GlobalStateReceiver(%d): ReadFrom failed: %v\n", port, err)
			continue
		}

		var globalState state.GlobalStateData
		if err := json.Unmarshal(receiveBuffer[:numBytes], &globalState); err != nil {
			fmt.Printf("broadcast.GlobalStateReceiver(%d): json.Unmarshal failed: %v\n", port, err)
			continue
		}

		incomingGlobalStateChannel <- globalState
	}
}
