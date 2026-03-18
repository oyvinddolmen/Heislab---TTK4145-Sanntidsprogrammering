package elevator

import (
	"heislab/elevator/elevIO"
	"heislab/management"
	"heislab/network"
	"heislab/orderManagement"
	"heislab/state"
)

// Polls sensors and buttons from elevIO and runs FSM
func RunElevator(
	elev *management.Elevator,
	globalState *state.GlobalState,
	elevChannels management.ElevChannels,
	networkChannels network.NetworkChannels,
) {
	go elevIO.PollFloorSensor(elevChannels.NewFloorChannel)
	go elevIO.PollObstructionSwitch(elevChannels.ObstructionChannel)
	go elevIO.PollButtons(elevChannels.ButtonPressChannel)
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
				if hallOrderAtFloor(elev, globalState) {
					orderManagement.ServeHallOrdersAtCurrentFloor(elev, globalState)
					setElevState(elev, globalState, management.ElevObstruction)
					continue
				}
				updateAssignmentsAndSetLights(elev, globalState)
				UpdateCurrentOrderAndDrive(elev, globalState)
			case button := <-elevChannels.ButtonPressChannel:
				if atButtonFloor(elev, button) {
					setElevState(elev, globalState, management.ElevObstruction)
					continue
				}
				registerAndBroadcastOrder(elev, globalState, button, networkChannels)
				updateAssignmentsAndSetLights(elev, globalState)
				if isCabOrderAtDifferentDir(elev) {
					clearOrdersAndBroadcast(elev, globalState, networkChannels)
					setElevState(elev, globalState, management.ElevObstruction)
					continue
				}
				UpdateCurrentOrderAndDrive(elev, globalState)
			case <-elevChannels.ObstructionChannel:
				setElevState(elev, globalState, management.ElevObstruction)
			case <-idleTimer.C: // Tells the elevator to take any order
				elev.SetLastOrderButtonType(elevIO.CabButton)
				UpdateCurrentOrderAndDrive(elev, globalState)
			}

		// ----------------- Case: MOVING -------------------------
		case management.ElevMoving:
			select {
			case <-networkChannels.WorldViewUpdateChannel:
				updateAssignmentsAndSetLights(elev, globalState)
			case floor := <-elevChannels.NewFloorChannel:
				elevIO.SetFloorIndicator(floor)
				elev.SetLastFloor(floor)
				elev.SetCanTakeOrders(true)
				resetCanTakeOrdersTimer()
				if shouldStop(elev, floor) {
					setMotorStop()
					elev.SetFloor(floor)
					ChooseDirectionAfterStop(elev, floor)
					clearOrdersAndBroadcast(elev, globalState, networkChannels)
					updateAssignmentsAndSetLights(elev, globalState)
					setElevState(elev, globalState, management.ElevObstruction)
				}
			case button := <-elevChannels.ButtonPressChannel:
				registerAndBroadcastOrder(elev, globalState, button, networkChannels)
				updateAssignmentsAndSetLights(elev, globalState)
			case <-elevChannels.ObstructionChannel:
				setElevState(elev, globalState, management.ElevObstruction)
			case <-canTakeOrdersTimer.C:
				elev.SetCanTakeOrders(false)
			}

		// ----------------- Case: OBSTRUCTION -------------------------
		case management.ElevObstruction:
			select {
			case <-networkChannels.WorldViewUpdateChannel:
				if hallOrderAtFloor(elev, globalState) {
					orderManagement.ServeHallOrdersAtCurrentFloor(elev, globalState)
					setElevState(elev, globalState, management.ElevObstruction)
					continue
				}
				updateAssignmentsAndSetLights(elev, globalState)
				updateCurrentOrderAndMoveDir(elev, globalState)
			case button := <-elevChannels.ButtonPressChannel:
				if atButtonFloor(elev, button) {
					startNewDoorTimer()
					continue
				}
				registerAndBroadcastOrder(elev, globalState, button, networkChannels)
				updateAssignmentsAndSetLights(elev, globalState)
				if isCabOrderAtDifferentDir(elev) {
					clearOrdersAndBroadcast(elev, globalState, networkChannels)
					startNewDoorTimer()
					continue
				}
				updateCurrentOrderAndMoveDir(elev, globalState)
			case <-doorTimer.C:
				if !elevIO.GetObstruction() {
					setElevState(elev, globalState, management.ElevIdle)
					continue
				}
				startNewDoorTimer()
			case obstructed := <-elevChannels.ObstructionChannel:
				if obstructed {
					doorTimer.Stop()
					elevIO.SetDoorOpenLamp(true)
				} else {
					startNewDoorTimer()
				}
			case <-canTakeOrdersTimer.C:
				elev.SetCanTakeOrders(false)
			}
		}
	}
}
