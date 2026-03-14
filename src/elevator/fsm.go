package elevator

import (
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/network"
	"heislab/orderManagement"
	"heislab/state"
	"time"
)

func RunElevator(
	elev *management.Elevator,
	globalState *state.GlobalState,
	elevChannels management.ElevChannels,
	networkChannels network.NetworkConnection,
) {
	go elevIO.PollFloorSensor(elevChannels.NewFloor)
	go elevIO.PollButtons(elevChannels.ButtonPresses)
	go elevIO.PollObstructionSwitch(elevChannels.Obstruction)
	setElevState(elev, globalState, management.ElevIdle)
	go runFSM(elev, globalState, elevChannels, networkChannels)
}

// -------------------------------------------------------------------------------------------
// FSM loop
// -------------------------------------------------------------------------------------------

func runFSM(
	elev *management.Elevator,
	globalState *state.GlobalState,
	elevChannels management.ElevChannels,
	networkChannels network.NetworkConnection,
) {
	for {
		switch elev.State {

		// ----------------- Case: IDLE -------------------------
		case management.ElevIdle:
			select {
			case <-networkChannels.WorldViewUpdateChannel:
				if needToOpenDoors(elev, globalState) {
					orderManagement.ServeHallRequestsAtCurrentFloor(elev, globalState)
					orderManagement.ClearOrdersAtCurrentFloor(elev, globalState)
					SetAllLights(elev, globalState) //Trenger denne?
					setElevState(elev, globalState, management.ElevObstruction)
				} else {
					updateAssignments(elev, globalState)
					UpdateCurrentOrderAndsafeDrive(elev, globalState)
				}
			case btn := <-elevChannels.ButtonPresses:
				HallOrderConflict := false	//Becomes true if HallUp and HallDown active at current floor and 
											//Caborder is pressed at a different direction then the current dirction
				OrderWasAtCurrentFloor := handleButtonPress(elev, globalState, btn, networkChannels)
				if !OrderWasAtCurrentFloor {
					if elev.GetFloor() != -1 {
						HallOrderConflict = orderManagement.ClearOrdersAtCurrentFloor(elev, globalState)
						updateAssignments(elev, globalState)
					}
					if HallOrderConflict {
						setElevState(elev, globalState, management.ElevObstruction)
					} else {
						UpdateCurrentOrderAndsafeDrive(elev, globalState)
					}
				}
			case <-elevChannels.Obstruction:
				setElevState(elev, globalState, management.ElevObstruction)
			case <-idleTimer.C:
				updateAssignments(elev, globalState)
				elev.LastOrder.ButtonType = elevIO.CabButton
				UpdateCurrentOrderAndsafeDrive(elev, globalState)
				if orderManagement.CurrentOrderPlaced(elev) == false {
					startIdleTimer()
				}
			}

		// ----------------- Case: MOVING -------------------------
		case management.ElevMoving:
			select {
			case <-networkChannels.WorldViewUpdateChannel:
				updateAssignments(elev, globalState)
			case floor := <-elevChannels.NewFloor:
				setFloorIndicator(floor)
				elev.SetElevLastFloor(floor)
				elev.SetElevCanTakeOrders(true)
				if ShouldStop(elev, floor) {
					setMotorStop()
					elev.SetElevFloor(floor)
					ChooseDirectionAfterStop(elev, floor)
					orderManagement.ClearOrdersAtCurrentFloor(elev, globalState)
					updateAssignments(elev, globalState)
					if orderManagement.CurrentOrderPlaced(elev) {
						orderManagement.UpdateCurrentOrder(elev, globalState)
						UpdateMoveDir(elev)
					}
					setElevState(elev, globalState, management.ElevObstruction)
				}
			case btn := <-elevChannels.ButtonPresses:
				handleButtonPress(elev, globalState, btn, networkChannels)
				orderManagement.UpdateCurrentOrder(elev, globalState)
			case <-elevChannels.Obstruction:
				setElevState(elev, globalState, management.ElevObstruction)
			case <-canTakeOrdersTimer.C:
				elev.SetElevCanTakeOrders(false)
			}

		// ----------------- Case: OBSTRUCTION -------------------------
		case management.ElevObstruction:
			select {
			case <-networkChannels.WorldViewUpdateChannel:
				if needToOpenDoors(elev, globalState) {
					orderManagement.ServeHallRequestsAtCurrentFloor(elev, globalState)
					orderManagement.ClearOrdersAtCurrentFloor(elev, globalState)
					SetAllLights(elev, globalState)
					startIdleTimer()
				} else {
					updateAssignments(elev, globalState)
					orderManagement.UpdateCurrentOrder(elev, globalState)
					UpdateMoveDir(elev)
				}
			case btn := <-elevChannels.ButtonPresses:
				mixedHallOrders := false
				OrderWasAtCurrentFloor := handleButtonPress(elev, globalState, btn, networkChannels)
				if !OrderWasAtCurrentFloor {

					if elev.GetFloor() != -1 {
						mixedHallOrders = orderManagement.ClearOrdersAtCurrentFloor(elev, globalState)
						updateAssignments(elev, globalState)
					}
					if mixedHallOrders || needToOpenDoors(elev, globalState) {
						doorTimer = time.NewTimer(doorOpenDuration)
					} else {
						orderManagement.UpdateCurrentOrder(elev, globalState)
						UpdateMoveDir(elev)
					}
				}
				if hallUpAndHallDownAndCabAtDifferentDir || needToOpenDoors(elev, globalState) {
					doorTimer = time.NewTimer(doorOpenDuration)
				} else {
					orderManagement.UpdateCurrentOrder(elev, globalState)
					UpdateMoveDir(elev)
				}
			case <-doorTimer.C:
				if !elevIO.GetObstruction() {
					setElevState(elev, globalState, management.ElevIdle)
				} else {
					startNewDoorTimer()
				}
			case <-canTakeOrdersTimer.C:
				elev.SetElevCanTakeOrders(false)

			}
		}
	}
}
