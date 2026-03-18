package elevator

import (
	"time"
)

// Timer variables
var doorTimer *time.Timer
var canTakeOrdersTimer *time.Timer // for detecting motor power loss in elev
var idleTimer *time.Timer          // safety measurement in case elevator don't drive to next order

// time durations
const doorOpenDuration = 3 * time.Second
const canTakeOrdersCountdown = 4 * time.Second
const IdleTimeOut = 2 * time.Second

func startNewDoorTimer() {
	if doorTimer != nil {
		doorTimer.Stop()
	}
	doorTimer = time.NewTimer(doorOpenDuration)
}

func startNewCanTakeOrdersTimer() {
	if canTakeOrdersTimer != nil {
		canTakeOrdersTimer.Stop()
	}
	canTakeOrdersTimer = time.NewTimer(canTakeOrdersCountdown)
}

func turnOffCanTakeOrdersTimer() {
	if canTakeOrdersTimer != nil {
		canTakeOrdersTimer.Stop()
	}
}

func resetCanTakeOrdersTimer() {
	if canTakeOrdersTimer != nil {
		canTakeOrdersTimer.Reset(canTakeOrdersCountdown)
	}
}

func startIdleTimer() {
	if idleTimer != nil {
		idleTimer.Stop()
	}
	idleTimer = time.NewTimer(IdleTimeOut)
}
