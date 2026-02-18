package network

// ---------------------------------------------------------------------------------------------------------------------
// Calling of communication functions in network-folder
// ---------------------------------------------------------------------------------------------------------------------

import (
	"fmt"
	"os"
	"time"

	"heislab/management"
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

type Config struct {
	PeerDiscoveryPort int // used by peers.Transmitter/Receiver (heartbeats)
	MessageBcastPort  int // used by bcast.Transmitter/Receiver (your actual data)
	NodeID            string
}

// Start launches goroutines for:
//  1. peer discovery (who is alive?)
//  2. broadcast messaging (sending/receiving typed messages)
//
// outgoingMessageChans: channels THIS node will broadcast OUT (bcast.Transmitter listens to them)
// incomingMessageChans: channels THIS node will receive IN from the network (bcast.Receiver writes to them)
//
// Returns:
//   - peerTxEnabled: send true/false to enable/disable announcing your presence
//   - peerUpdates: stream of PeerUpdate (New/Lost/Peers)
//   - myID: the node ID used on the network
func InitNetwork(
	cfg Config,
	outgoingMessageChans []interface{},
	incomingMessageChans []interface{},
) (peerTxEnabled chan<- bool, peerUpdates <-chan peers.PeerUpdate, myID string) {

	myID = cfg.NodeID
	if myID == "" {
		ip, err := localip.LocalIP()
		if err != nil {
			myID = fmt.Sprintf("unknown-%d", os.Getpid())
		} else {
			myID = fmt.Sprintf("%s-%d", ip, os.Getpid())
		}
	}

	// Channel to turn peer announcements on/off (useful for simulating disconnect).
	peerTxEnabledCh := make(chan bool, 1)
	peerTxEnabledCh <- true

	// Channel where the peers.Receiver posts updates about peers.
	peerUpdatesCh := make(chan peers.PeerUpdate, 16)

	// ---- Peer discovery: "I'm alive" beacons + tracking others ----
	go peers.Transmitter(cfg.PeerDiscoveryPort, myID, peerTxEnabledCh)
	go peers.Receiver(cfg.PeerDiscoveryPort, peerUpdatesCh)

	// ---- Message broadcast: your actual elevator messages ----
	if len(outgoingMessageChans) > 0 {
		go bcast.Transmitter(cfg.MessageBcastPort, outgoingMessageChans...)
	}
	if len(incomingMessageChans) > 0 {
		go bcast.Receiver(cfg.MessageBcastPort, incomingMessageChans...)
	}

	return peerTxEnabledCh, peerUpdatesCh, myID
}
