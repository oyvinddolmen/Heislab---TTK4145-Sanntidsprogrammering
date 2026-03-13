package elevator

import (
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/network"
	"heislab/orderManagement"
	"heislab/state"
	"time"
)

type ElevChannels struct {
	NewFloor    chan int
	Obstruction chan bool
	StopBtn     chan bool
	BtnPresses  chan elevIO.ButtonEvent
}



func RunElevator(
	elev *management.Elevator,
	gs *state.GlobalState,
	elevChannels ElevChannels,
	networkChannels network.NetworkConn,
) {
	go elevIO.PollFloorSensor(elevChannels.NewFloor)
	go elevIO.PollButtons(elevChannels.BtnPresses)
	go elevIO.PollStopButton(elevChannels.StopBtn)
	go elevIO.PollObstructionSwitch(elevChannels.Obstruction)
	setElevState(elev, gs, management.ElevIdle)
	go runFSM(elev, gs, elevChannels, networkChannels)
}

// -------------------------------------------------------------------------------------------
// FSM loop
// -------------------------------------------------------------------------------------------

func runFSM(
	elev *management.Elevator,
	gs *state.GlobalState,
	elevChannels ElevChannels,
	networkChannels network.NetworkConn,
) {
	for {
		switch elev.State {

		// ----------------- Case: IDLE -------------------------
		case management.ElevIdle:
			select {
			case <-networkChannels.WorldViewUpdate:
				if needToOpenDoors(elev, gs) {
					orderManagement.ServeHallRequestsAtCurrentFloor(elev, gs)
					orderManagement.ClearOrdersAndTurnOffLights(elev, gs)
					SetAllLights(elev, gs)
					setElevState(elev, gs, management.ElevObstruction)
				} else {
					updateAssignments(elev, gs)
					UpdateCurrentOrderAndsafeDrive(elev, gs)
				}
			case <-idleTimer.C:
				elev.LastOrder.ButtonType = elevIO.CabButton
				UpdateCurrentOrderAndsafeDrive(elev, gs)
				if elev.CurrentOrder.OrderPlaced == false {
					startIdleTimer()
				}
			case <-elevChannels.Obstruction:
				setElevState(elev, gs, management.ElevObstruction)
			case <-elevChannels.StopBtn:
				if elev.Floor != -1 {
					setElevState(elev, gs, management.ElevObstruction)
				} else {
					setElevState(elev, gs, management.ElevStop)
				}
			case btn := <-elevChannels.BtnPresses:
				HallUpAndHallDownAndCabAtDifferentDir := false
				OrderWasAtCurrentFloor := handleButtonPress(elev, gs, btn, networkChannels)
				if !OrderWasAtCurrentFloor {
					if elev.GetFloor() != -1 {
						HallUpAndHallDownAndCabAtDifferentDir = orderManagement.ClearOrdersAndTurnOffLights(elev, gs)
						updateAssignments(elev, gs)
					}
					if HallUpAndHallDownAndCabAtDifferentDir {
						setElevState(elev, gs, management.ElevObstruction)
					} else {
						UpdateCurrentOrderAndsafeDrive(elev, gs)
					}
				}
			}

		// ----------------- Case: MOVING -------------------------
		case management.ElevMoving:
			select {
			case <-networkChannels.WorldViewUpdate:
				updateAssignments(elev, gs)
			case floor := <-elevChannels.NewFloor:
				setFloorIndicator(floor)
				elev.SetElevLastFloor(floor)
				elev.SetElevCanTakeOrders(true)
				if ShouldStop(elev, floor) {
					setMotorStop()
					elev.SetElevFloor(floor)
					ChooseDirectionAfterStop(elev, floor)
					orderManagement.ClearOrdersAndTurnOffLights(elev, gs)
					updateAssignments(elev, gs)
					if elev.CurrentOrder.OrderPlaced == false {
						orderManagement.UpdateCurrentOrder(elev, gs)
						UpdateMoveDir(elev)
					}
					setElevState(elev, gs, management.ElevObstruction)
				}
			case <-elevChannels.Obstruction:
				setElevState(elev, gs, management.ElevObstruction)
			case <-elevChannels.StopBtn:
				setElevState(elev, gs, management.ElevStop)
			case btn := <-elevChannels.BtnPresses:
				handleButtonPress(elev, gs, btn, networkChannels)
				orderManagement.UpdateCurrentOrder(elev, gs)
			case <-canTakeOrdersTimer.C:
				elev.SetElevCanTakeOrders(false)
			}

		// ----------------- Case: OBSTRUCTION -------------------------
		case management.ElevObstruction:
			select {
			case <-networkChannels.WorldViewUpdate:
				if needToOpenDoors(elev, gs) {
					orderManagement.ServeHallRequestsAtCurrentFloor(elev, gs)
					orderManagement.ClearOrdersAndTurnOffLights(elev, gs)
					SetAllLights(elev, gs)
					startIdleTimer()
				} else {
					updateAssignments(elev, gs)
					orderManagement.UpdateCurrentOrder(elev, gs)
					UpdateMoveDir(elev)
				}
			case <-doorTimer.C:
				if !elevIO.GetObstruction() {
					setElevState(elev, gs, management.ElevIdle)
				} else {
					startNewDoorTimer() // eller set elev state obstruction??
				}
			case <-canTakeOrdersTimer.C:
				elev.SetElevCanTakeOrders(false)
			case btn := <-elevChannels.BtnPresses:
				mixedHallOrders := false
				OrderWasAtCurrentFloor := handleButtonPress(elev, gs, btn, networkChannels)
				if !OrderWasAtCurrentFloor {

					if elev.GetFloor() != -1 {
						mixedHallOrders = orderManagement.ClearOrdersAndTurnOffLights(elev, gs)
						updateAssignments(elev, gs)
					}
					if mixedHallOrders || needToOpenDoors(elev, gs) {
						doorTimer = time.NewTimer(doorOpenDuration)
					} else {
						orderManagement.UpdateCurrentOrder(elev, gs)
						UpdateMoveDir(elev)
					}
				}
				if hallUpAndHallDownAndCabAtDifferentDir || needToOpenDoors(elev, gs) {
					doorTimer = time.NewTimer(doorOpenDuration)
				} else {
					orderManagement.UpdateCurrentOrder(elev, gs)
					UpdateMoveDir(elev)
				}
			}
		}
	}
}
