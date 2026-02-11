package network

import (
	"heislab/elevator"
	"time"
)

type NetworkChannels struct {
	RcvChannel   chan elevator.Elevator
	BcastChannel chan elevator.Elevator
}

func RunNetwork(channels NetworkChannels) {
	// Funksjonen er bare en placeholder. Bare å endre navn og det den gjør
}

func BcastElevInfo(BcastChannel chan elevator.Elevator) {
	time.Sleep(2 * time.Millisecond)
	BcastChannel <- elevator.Elev
}
