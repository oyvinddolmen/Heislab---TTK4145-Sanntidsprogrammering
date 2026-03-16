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
	networkChannels network.NetworkChannels,
) {
	go elevIO.PollFloorSensor(elevChannels.NewFloorChannel)
	go elevIO.PollButtons(elevChannels.ButtonPressChannel)
	go elevIO.PollObstructionSwitch(elevChannels.ObstructionChannel)
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
	networkChannels network.NetworkChannels,
) {
	for {
		switch elev.GetState() {

		// ----------------- Case: IDLE -------------------------
		case management.ElevIdle:
			select {
			case <-networkChannels.WorldViewUpdateChannel:
				if needToOpenDoors(elev, globalState) {
					orderManagement.ServeHallOrdersAtCurrentFloor(elev, globalState)
					orderManagement.ClearOrdersAtCurrentFloor(elev, globalState)
					SetAllLights(elev, globalState) //Trenger denne?
					setElevState(elev, globalState, management.ElevObstruction)
				} else {
					updateAssignments(elev, globalState)
					UpdateCurrentOrderAndsafeDrive(elev, globalState)
				}
			case button := <-elevChannels.ButtonPressChannel:
				HallOrderConflict := false //Becomes true if HallUp and HallDown active at current floor and
				//Caborder is pressed at a different direction then the current dirction
				OrderWasAtCurrentFloor := handleButtonPress(elev, globalState, button, networkChannels)
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
			case <-elevChannels.ObstructionChannel:
				setElevState(elev, globalState, management.ElevObstruction)
			case <-idleTimer.C:
				updateAssignments(elev, globalState)
				elev.SetLastOrderButtonType(elevIO.CabButton)
				UpdateCurrentOrderAndsafeDrive(elev, globalState)
				if !elev.GetCurrentOrderActiveStatus() {
					startIdleTimer()
				}
			}

		// ----------------- Case: MOVING -------------------------
		case management.ElevMoving:
			select {
			case <-networkChannels.WorldViewUpdateChannel:
				updateAssignments(elev, globalState)
			case floor := <-elevChannels.NewFloorChannel:
				setFloorIndicator(floor)
				elev.SetLastFloor(floor)
				elev.SetCanTakeOrders(true)
				if ShouldStop(elev, floor) {
					setMotorStop()
					elev.SetFloor(floor)
					ChooseDirectionAfterStop(elev, floor)
					orderManagement.ClearOrdersAtCurrentFloor(elev, globalState)
					updateAssignments(elev, globalState)
					if elev.GetCurrentOrderActiveStatus() {
						orderManagement.UpdateCurrentOrder(elev, globalState)
						UpdateMoveDir(elev)
					}
					setElevState(elev, globalState, management.ElevObstruction)
				}
			case button := <-elevChannels.ButtonPressChannel:
				handleButtonPress(elev, globalState, button, networkChannels)
				orderManagement.UpdateCurrentOrder(elev, globalState)
			case <-elevChannels.ObstructionChannel:
				setElevState(elev, globalState, management.ElevObstruction)
			case <-canTakeOrdersTimer.C:
				elev.SetCanTakeOrders(false)
			}

		// ----------------- Case: OBSTRUCTION -------------------------
		case management.ElevObstruction:
			select {
			case <-networkChannels.WorldViewUpdateChannel:
				if needToOpenDoors(elev, globalState) {
					orderManagement.ServeHallOrdersAtCurrentFloor(elev, globalState)
					orderManagement.ClearOrdersAtCurrentFloor(elev, globalState)
					SetAllLights(elev, globalState)
					startIdleTimer()
				} else {
					updateAssignments(elev, globalState)
					orderManagement.UpdateCurrentOrder(elev, globalState)
					UpdateMoveDir(elev)
				}
			case button := <-elevChannels.ButtonPressChannel:
				mixedHallOrders := false
				OrderWasAtCurrentFloor := handleButtonPress(elev, globalState, button, networkChannels)
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
				elev.SetCanTakeOrders(false)
			}
		}
	}
}
