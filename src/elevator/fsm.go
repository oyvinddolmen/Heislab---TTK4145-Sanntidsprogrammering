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
					orderManagement.ClearOrdersAtCurrentFloor(elev, gs)
					SetAllLights(elev, gs) //Trenger denne?
					setElevState(elev, gs, management.ElevObstruction)
				} else {
					updateAssignments(elev, gs)
					UpdateCurrentOrderAndsafeDrive(elev, gs)
				}
			case btn := <-elevChannels.BtnPresses:
				HallOrderConflict := false	//Becomes true if HallUp and HallDown active at current floor and 
											//Caborder is pressed at a different direction then the current dirction
				OrderWasAtCurrentFloor := handleButtonPress(elev, gs, btn, networkChannels)
				if !OrderWasAtCurrentFloor {
					if elev.GetFloor() != -1 {
						HallOrderConflict = orderManagement.ClearOrdersAtCurrentFloor(elev, gs)
						updateAssignments(elev, gs)
					}
					if HallOrderConflict {
						setElevState(elev, gs, management.ElevObstruction)
					} else {
						UpdateCurrentOrderAndsafeDrive(elev, gs)
					}
				}
			case <-elevChannels.Obstruction:
				setElevState(elev, gs, management.ElevObstruction)
			case <-idleTimer.C:
				elev.LastOrder.ButtonType = elevIO.CabButton
				UpdateCurrentOrderAndsafeDrive(elev, gs)
				if orderManagement.CurrentOrderPlaced(elev) == false {
					startIdleTimer()
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
					orderManagement.ClearOrdersAtCurrentFloor(elev, gs)
					updateAssignments(elev, gs)
					if orderManagement.CurrentOrderPlaced(elev) {
						orderManagement.UpdateCurrentOrder(elev, gs)
						UpdateMoveDir(elev)
					}
					setElevState(elev, gs, management.ElevObstruction)
				}
			case btn := <-elevChannels.BtnPresses:
				handleButtonPress(elev, gs, btn, networkChannels)
				orderManagement.UpdateCurrentOrder(elev, gs)
			case <-elevChannels.Obstruction:
				setElevState(elev, gs, management.ElevObstruction)
			case <-canTakeOrdersTimer.C:
				elev.SetElevCanTakeOrders(false)
			}

		// ----------------- Case: OBSTRUCTION -------------------------
		case management.ElevObstruction:
			select {
			case <-networkChannels.WorldViewUpdate:
				if needToOpenDoors(elev, gs) {
					orderManagement.ServeHallRequestsAtCurrentFloor(elev, gs)
					orderManagement.ClearOrdersAtCurrentFloor(elev, gs)
					SetAllLights(elev, gs)
					startIdleTimer()
				} else {
					updateAssignments(elev, gs)
					orderManagement.UpdateCurrentOrder(elev, gs)
					UpdateMoveDir(elev)
				}
			case btn := <-elevChannels.BtnPresses:
				mixedHallOrders := false
				OrderWasAtCurrentFloor := handleButtonPress(elev, gs, btn, networkChannels)
				if !OrderWasAtCurrentFloor {

					if elev.GetFloor() != -1 {
						mixedHallOrders = orderManagement.ClearOrdersAtCurrentFloor(elev, gs)
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
			case <-doorTimer.C:
				if !elevIO.GetObstruction() {
					setElevState(elev, gs, management.ElevIdle)
				} else {
					startNewDoorTimer()
				}
			case <-canTakeOrdersTimer.C:
				elev.SetElevCanTakeOrders(false)

			}
		}
	}
}
