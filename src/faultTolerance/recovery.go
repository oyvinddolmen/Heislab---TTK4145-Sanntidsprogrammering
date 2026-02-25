package faultTolerance

import (
	"heislab/elevio"
	"heislab/management"
	"heislab/orderManagement"
	"fmt"
	"time"
)

// Called once when elevator boots

func RecoverOnStartup(rx <-chan orderManagement.GlobalStateType) {

    timeout := time.After(1 * time.Second)

    // Vent på eksisterende GlobalState 
    for {
        select {

        case gs := <-rx:
            orderManagement.GlobalStateMutex.Lock()
            orderManagement.GlobalState = gs
            orderManagement.GlobalStateMutex.Unlock()
            goto RECOVER

        case <-timeout:
            fmt.Println("No GlobalState received on startup, starting fresh")
            goto RECOVER
        }
    }

RECOVER:

    localIP := management.Elev.IP

    orderManagement.GlobalStateMutex.Lock()
    defer orderManagement.GlobalStateMutex.Unlock()

    // Gjenopprett cab-orders hvis de fantes fra før 
    oldState, exists := orderManagement.GlobalState.States[localIP]
    if exists {
        for f := 0; f < management.NumFloors; f++ {
            if oldState.CabRequests[f] {
                management.Elev.Orders[f][elevio.BT_Cab].OrderPlaced = true
                management.Elev.Orders[f][elevio.BT_Cab].Finished = false
                management.Elev.Orders[f][elevio.BT_Cab].ElevIP = management.Elev.IP
            }
        }
    }

    // Registrer oss selv i GlobalState 
    orderManagement.GlobalState.States[localIP] = orderManagement.ConvertElevatorToJSON(management.Elev)
}

