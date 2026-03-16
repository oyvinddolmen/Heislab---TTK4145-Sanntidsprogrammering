package elevIO

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const pollRate = 20 * time.Millisecond
const numButtonTypes int = 3

var initialized bool = false
var numFloors int = 4
var mutex sync.Mutex
var connection net.Conn

type MotorDirection int
const (
	MotorDirUp   MotorDirection = 1
	MotorDirDown MotorDirection = -1
	MotorDirStop MotorDirection = 0
)

type ButtonType int
const (
	HallUpButton   ButtonType = 0
	HallDownButton ButtonType = 1
	CabButton      ButtonType = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

// TODO: struct for cab and floor lights brukes heller ikke?
type CabFloorLights struct {
	Floor  int
	Button ButtonType
	Value  bool
}

func InitElevatorIO(elevAddress string, totalFloors int) {
	if initialized {
		fmt.Println("Driver already initialized!")
		return
	}

	numFloors = totalFloors
	mutex = sync.Mutex{}
	var err error
	connection, err = net.Dial("tcp", elevAddress)
	if err != nil {
		panic(err.Error())
	}

	initialized = true
}

func SetMotorDirection(direction MotorDirection) {
	write([4]byte{1, byte(direction), 0, 0})
}

func SetButtonLamp(button ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

func PollButtons(receiver chan<- ButtonEvent) {
	previousButtonState := make([][3]bool, numFloors)

	for {
		time.Sleep(pollRate)
		for floor := 0; floor < numFloors; floor++ {
			for button := ButtonType(0); button < ButtonType(numButtonTypes); button++ {
				buttonState := GetButton(button, floor)

				if buttonState != previousButtonState[floor][button] && buttonState != false {
					receiver <- ButtonEvent{floor, ButtonType(button)}
				}

				previousButtonState[floor][button] = buttonState
			}
		}
	}
}

// TODO: Brukes denne?
func PollFloorSensor(receiver chan<- int) {
	previousFloor := -1

	for {
		time.Sleep(pollRate)
		currentFloor := GetFloor()
		if currentFloor != previousFloor && currentFloor != -1 {
			receiver <- currentFloor
		}
		previousFloor = currentFloor
	}
}

func PollStopButton(receiver chan<- bool) {
	previousStopPressed := false

	for {
		time.Sleep(pollRate)
		currentStopPressed := GetStop()
		if currentStopPressed != previousStopPressed {
			receiver <- currentStopPressed
		}
		previousStopPressed = currentStopPressed
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	previousObstructionState := false

	for {
		time.Sleep(pollRate)
		currentObstructionState := GetObstruction()
		if currentObstructionState != previousObstructionState {
			receiver <- currentObstructionState
		}
		previousObstructionState = currentObstructionState
	}
}

func GetButton(button ButtonType, floor int) bool {
	response := read([4]byte{6, byte(button), byte(floor), 0})
	buttonPressed := toBool(response[1])
	return buttonPressed
}

func GetFloor() int {
	response := read([4]byte{7, 0, 0, 0})
	floorSensorActive := response[1]
	floor := int(response[2])

	if floorSensorActive != 0 {
		return floor
	}
	return -1
}

func GetStop() bool {
	response := read([4]byte{8, 0, 0, 0})
	stopPressed := toBool(response[1])
	return stopPressed
}

func GetObstruction() bool {
	response := read([4]byte{9, 0, 0, 0})
	obstructionState := toBool(response[1])
	return obstructionState
}

func read(input [4]byte) [4]byte {
	mutex.Lock()
	defer mutex.Unlock()

	_, err := connection.Write(input[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var output [4]byte
	_, err = connection.Read(output[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return output
}

func write(input [4]byte) {
	mutex.Lock()
	defer mutex.Unlock()

	_, err := connection.Write(input[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(inputBool bool) byte {
	var outputByte byte = 0
	if inputBool {
		outputByte = 1
	}
	return outputByte
}

func toBool(inputByte byte) bool {
	var outputBool bool = false
	if inputByte != 0 {
		outputBool = true
	}
	return outputBool
}
