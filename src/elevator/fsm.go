package elevator

type Elevator struct {
	State  ElevState
	orders OrderMatrix
	lights LightMatrix
}

type ElevState struct {
	ID        int
	direction Direction
	floor     int
	available bool
}

// placeholders
type OrderMatrix struct {
	placeholder int
}
type LightMatrix struct {
	placeholder2 int
}

