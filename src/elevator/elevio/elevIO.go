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

// Turns button lamp on or off for the specified button and floor.
func SetButtonLamp(button ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

// Turns floor indicator lamp on for specified floor.
func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

// Turns door open lamp on or off.
func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

// Turns stop lamp on or off.
func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

// Constantly checks button states and sends changes on channel.
func PollButtons(buttonPressChannel chan<- ButtonEvent) {
	previousButtonState := make([][3]bool, numFloors)

	for {
		time.Sleep(pollRate)
		for floor := 0; floor < numFloors; floor++ {
			for button := ButtonType(0); button < ButtonType(numButtonTypes); button++ {
				buttonState := getButtonState(button, floor)

				if buttonState != previousButtonState[floor][button] && buttonState != false {
					buttonPressChannel <- ButtonEvent{floor, ButtonType(button)}
				}

				previousButtonState[floor][button] = buttonState
			}
		}
	}
}

// Constantly checks the elevator's current floor and sends changes on channel.
func PollFloorSensor(newFloorChannel chan<- int) {
	previousFloor := -1

	for {
		time.Sleep(pollRate)
		currentFloor := GetFloor()
		if currentFloor != previousFloor && currentFloor != -1 {
			newFloorChannel <- currentFloor
		}
		previousFloor = currentFloor
	}
}

// Constantly checks state of the obstruction switch and sends changes on channel.
func PollObstructionSwitch(obstructionChannel chan<- bool) {
	previousObstructionState := false

	for {
		time.Sleep(pollRate)
		currentObstructionState := GetObstruction()
		if currentObstructionState != previousObstructionState {
			obstructionChannel <- currentObstructionState
		}
		previousObstructionState = currentObstructionState
	}
}

// Checks if specified button is pressed and returns true if it is.
func getButtonState(button ButtonType, floor int) bool {
	response := read([4]byte{6, byte(button), byte(floor), 0})
	buttonPressed := toBool(response[1])
	return buttonPressed
}

// Checks floor sensor status and returns floor number if at a floor, -1 else.
func GetFloor() int {
	response := read([4]byte{7, 0, 0, 0})
	floorSensorActive := toBool(response[1])
	floor := int(response[2])

	if floorSensorActive {
		return floor
	}
	return -1
}

// Reads and returns obstruction switch state.
func GetObstruction() bool {
	response := read([4]byte{9, 0, 0, 0})
	obstructionState := toBool(response[1])
	return obstructionState
}

// Sends input command to elevator simulator and returns response.
func read(input [4]byte) [4]byte {
	mutex.Lock()
	defer mutex.Unlock()

	_, err := connection.Write(input[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var response [4]byte
	_, err = connection.Read(response[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return response
}

// Sends input command to elevator server.
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
