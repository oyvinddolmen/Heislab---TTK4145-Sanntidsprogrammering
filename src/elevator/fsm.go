package elevator

import (
	"fmt"
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
				if needToOpenDoors(elev, globalState) {
					orderManagement.ServeHallOrdersAtCurrentFloor(elev, globalState) 
					setElevState(elev, globalState, management.ElevObstruction)
				} else {
					updateAssignments(elev, globalState)
					UpdateCurrentOrderAndDrive(elev, globalState)
				}
			case button := <-elevChannels.ButtonPressChannel:
				if atButtonFloor(elev, button){
					setElevState(elev, globalState, management.ElevObstruction)
					continue
				}
				registerOrder(elev, globalState, button)
				network.SendGlobalState(elev, globalState, networkChannels.OutgoingGlobalStateChannel)
				orderManagement.RunHallAssignerAndApplyAssignments(elev, globalState)
				SetAllLights(elev, globalState)
				if elev.IsAtFloor() {
					orderManagement.ClearOrdersAtCurrentFloor(elev, globalState)
					updateAssignments(elev, globalState)
				}
				if isCabOrderAtDifferentDir(elev) {
					setElevState(elev, globalState, management.ElevObstruction)
				} else {
					UpdateCurrentOrderAndDrive(elev, globalState)
				}
			case <-elevChannels.ObstructionChannel:
				setElevState(elev, globalState, management.ElevObstruction)
			case <-idleTimer.C:
				updateAssignments(elev, globalState)
				elev.SetLastOrderButtonType(elevIO.CabButton)
				UpdateCurrentOrderAndDrive(elev, globalState)
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
				resetCanTakeOrdersTimer()
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
				registerOrder(elev, globalState, button)
				network.SendGlobalState(elev, globalState, networkChannels.OutgoingGlobalStateChannel)
				updateAssignments(elev, globalState)
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
					setElevState(elev, globalState, management.ElevObstruction)
				} else {
					updateAssignments(elev, globalState)
					updateCurrentOrderAndMoveDir(elev, globalState)
				}
			case button := <-elevChannels.ButtonPressChannel:
				if atButtonFloor(elev, button){
					setElevState(elev, globalState, management.ElevObstruction)
					continue
				}
				registerOrder(elev, globalState, button)
				network.SendGlobalState(elev, globalState, networkChannels.OutgoingGlobalStateChannel)
				updateAssignments(elev, globalState)
				if elev.IsAtFloor() {
					orderManagement.ClearOrdersAtCurrentFloor(elev, globalState)
					updateAssignments(elev, globalState)
				}
				if isCabOrderAtDifferentDir(elev) || needToOpenDoors(elev, globalState) {
					doorTimer = time.NewTimer(doorOpenDuration)
				} 
				updateCurrentOrderAndMoveDir(elev, globalState)
				
			case <-doorTimer.C:
				if !elevIO.GetObstruction() {
					fmt.Println("doorTimer!!!!!!")
					setElevState(elev, globalState, management.ElevIdle)
					continue
				} 
				startNewDoorTimer()
				
			case <-canTakeOrdersTimer.C:
				elev.SetCanTakeOrders(false)
			}
		}
	}
}
