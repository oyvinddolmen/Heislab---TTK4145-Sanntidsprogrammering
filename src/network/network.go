package network

// ---------------------------------------------------------------------------------------------------------------------
// Calling of communication functions in network-folder
// ---------------------------------------------------------------------------------------------------------------------


import (
	"heislab/management"
	"time"
)

type NetworkChannels struct {
	RcvChannel   chan management.Elevator
	BcastChannel chan management.Elevator
}

func RunNetwork(channels NetworkChannels) {
	// Funksjonen er bare en placeholder. Bare å endre navn og det den gjør
}

func BcastElevInfo(BcastChannel chan management.Elevator) {
	time.Sleep(2 * time.Millisecond)
	BcastChannel <- management.Elev
}
