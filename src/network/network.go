package network

import(
	"heislab/orderManagment"
	"heislab/elevator"
)

type ElevTransferData struct {
	Id           int
	Floor        int
	CurrentOrder orderManagment.Order
	State        int
	Orders       [elevator.NumFloors][elevator.NumButtons]orderManagment.Order
}

type NetworkChannels struct {
	RcvChannel   chan ElevTransferData
	BcastChannel chan ElevTransferData
}

func RunNetwork(channels NetworkChannels) {
	// Funksjonen er bare en placeholder. Bare å endre navn og det den gjør
}