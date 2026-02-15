package network

// ---------------------------------------------------------------------------------------------------------------------
// Calling of communication functions in network-folder
// ---------------------------------------------------------------------------------------------------------------------


import (
	"heislab/managment"
	"time"
)

type NetworkChannels struct {
	RcvChannel   chan managment.Elevator
	BcastChannel chan managment.Elevator
}

func RunNetwork(channels NetworkChannels) {
	// Funksjonen er bare en placeholder. Bare å endre navn og det den gjør
}

func BcastElevInfo(BcastChannel chan managment.Elevator) {
	time.Sleep(2 * time.Millisecond)
	BcastChannel <- managment.Elev
}
