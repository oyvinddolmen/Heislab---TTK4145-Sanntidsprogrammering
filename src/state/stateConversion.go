package state

import (
	"heislab/management"
)

type ElevatorStateJSON struct {
	Behavior      string `json:"behaviour"` // idle, moving, doorOpen, offline
	Floor         int    `json:"floor"`
	Direction     string `json:"direction"`
	CabOrders     []bool `json:"cabRequests"`
	CanTakeOrders bool   `json:"canTakeOrders"`
}

func ConvertElevatorToJSON(elev *management.Elevator) ElevatorStateJSON {
	cabOrders := make([]bool, management.NumFloors)
	for floor := 0; floor < management.NumFloors; floor++ {
		cabOrders[floor] = elev.GetOrderActiveStatus(floor, management.CabButton)
	}

	return ElevatorStateJSON{
		Behavior:      ConvertState(elev.GetState()),
		Floor:         elev.GetLastFloor(),
		Direction:     convertDirection(elev.GetMoveDir()),
		CabOrders:     cabOrders,
		CanTakeOrders: elev.GetCanTakeOrders(),
	}
}

func ConvertState(state management.State) string {
	switch state {
	case management.ElevIdle:
		return "idle"
	case management.ElevMoving:
		return "moving"
	case management.ElevInit:
		return "moving"
	default:
		return "idle"
	}
}

func convertDirection(direction management.Direction) string {
	switch direction {
	case management.DirUp:
		return "up"
	case management.DirDown:
		return "down"
	default:
		return "stop"
	}
}
