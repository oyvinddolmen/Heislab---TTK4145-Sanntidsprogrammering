package elevator

import (
	"fmt"
	"heislab/elevio"
	"heislab/management"
	"heislab/network"
	"heislab/orderManagement"
	"time"
)

// -------------------------------------------------------------------------------------------
// Timer for door
// -------------------------------------------------------------------------------------------
var doorTimer *time.Timer

const doorOpenDuration = 2 * time.Second

// -------------------------------------------------------------------------------------------
// Initialize FSM
// -------------------------------------------------------------------------------------------
func InitFSM(elevID string, NumFloors int) {
	noOrder := management.Order{Floor: -1, ButtonType: -1, ElevID: ""}

	management.Elev.ID = elevID
	management.Elev.State = management.ElevInit
	management.Elev.Floor = -1 // Between floors
	management.Elev.LastFloor = 0
	management.Elev.MoveDir = management.DirIdle
	management.Elev.CurrentOrder = noOrder

	for floor := 0; floor < NumFloors; floor++ {
		for button := 0; button < management.NumButtons; button++ {
			management.Elev.Orders[floor][button] = management.Order{
				Floor:       floor,
				ButtonType:  elevio.ButtonType(button),
				ElevID:      "",
				OrderPlaced: false,
			}
		}
	}
}

// -------------------------------------------------------------------------------------------
// Run FSM
// -------------------------------------------------------------------------------------------
func RunElevator(
	gs *orderManagement.GlobalState,
	elevChannels management.ElevChannels,
	networkChannels network.NetworkConn,
) {
	go elevio.PollFloorSensor(elevChannels.LastFloor)
	go elevio.PollButtons(elevChannels.BtnPresses)
	go elevio.PollStopButton(elevChannels.StopBtn)
	go elevio.PollObstructionSwitch(elevChannels.Obstruction)

	go runFSM(gs, elevChannels, networkChannels)
}

// -------------------------------------------------------------------------------------------
// FSM loop
// -------------------------------------------------------------------------------------------
func runFSM(
	gs *orderManagement.GlobalState,
	elevChannels management.ElevChannels,
	networkChannels network.NetworkConn,
) {
	for {
		switch management.Elev.State {

		// ----------------- Case: IDLE -------------------------
		case management.ElevIdle:
			select {
			case <-networkChannels.WorldViewUpdate:
				orderManagement.RunHallAssigner(gs)
				setHallLightOnAllPanels(gs)
				safeDrive()
			case floor := <-elevChannels.LastFloor:
				updateFloor(floor)
			case <-elevChannels.Obstruction:
				setElevState(gs, management.ElevObstruction)
			case <-elevChannels.StopBtn:
				if management.Elev.Floor != -1 {
					setElevState(gs, management.ElevObstruction)
				} else {
					setElevState(gs, management.ElevStop)
				}
			case btn := <-elevChannels.BtnPresses:
				handleButtonPress(gs, btn, networkChannels)
				safeDrive()
			}

		// ----------------- Case: MOVING -------------------------
		case management.ElevMoving:
			select {
			case <-networkChannels.WorldViewUpdate:
				orderManagement.RunHallAssigner(gs)
				setHallLightOnAllPanels(gs)
				safeDrive()
			case floor := <-elevChannels.LastFloor:
				updateFloor(floor)
				if orderManagement.ShouldStop(&management.Elev) {
					stopElevator()
					orderManagement.ClearOrdersAndTurnOfLights(gs)
					gs.Print()
					orderManagement.RunHallAssigner(gs)
					setElevState(gs, management.ElevObstruction)
				}
			case <-elevChannels.StopBtn:
				setElevState(gs, management.ElevStop)
			case btn := <-elevChannels.BtnPresses:
				handleButtonPress(gs, btn, networkChannels)
				safeDrive()
			}

		// ----------------- Case: STOP -------------------------
		case management.ElevStop:
			select {
			case <-networkChannels.WorldViewUpdate:
				orderManagement.RunHallAssigner(gs)
				setHallLightOnAllPanels(gs)
			case btn := <-elevChannels.BtnPresses:
				handleButtonPress(gs, btn, networkChannels)
			case <-elevChannels.StopBtn:
				if getFloor() != -1 {
					setElevState(gs, management.ElevObstruction)
				} else {
					setElevState(gs, management.ElevIdle)
				}
			}

		// ----------------- Case: OBSTRUCTION -------------------------
		case management.ElevObstruction:
			select {
			case <-networkChannels.WorldViewUpdate:
				orderManagement.RunHallAssigner(gs)
				setHallLightOnAllPanels(gs)
			case <-doorTimer.C:
				if !elevio.GetObstruction() {
					elevio.SetDoorOpenLamp(false)
					fmt.Println("Nåværende floor: ", management.Elev.Floor, "og state ", management.Elev.State)
					fmt.Println("Går til idle nå")
					setElevState(gs, management.ElevIdle)
				}
			case btn := <-elevChannels.BtnPresses:
				handleButtonPress(gs, btn, networkChannels)
			}
		}
	}
}

// -------------------------------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------------------------------
func handleButtonPress(gs *orderManagement.GlobalState, btn elevio.ButtonEvent, networkChannels network.NetworkConn) {
	order := orderManagement.CreateOrder(btn)

	if order.Floor == management.Elev.Floor {
		setElevState(gs, management.ElevObstruction)
		return
	}

	if order.ButtonType == management.CabButton {
		orderManagement.AddOrderToOrders(order)
		gs.UpdateLocalGlobalState()
	} else {
		gs.AddHallRequest(order)
		gs.IncrementHallRequestVersion(order.Floor, order.ButtonType)
	}

	network.SendGlobalState(gs, networkChannels.GlobalStateTx)
	orderManagement.RunHallAssigner(gs)
	elevio.SetButtonLamp(btn.Button, btn.Floor, true)
}

func setMoveDir(moveDir management.Direction) {
	management.Elev.MoveDir = moveDir
}

func setElevLastFloor(lastFloor int) {
	management.Elev.LastFloor = lastFloor
}

func setElevFloor(floor int) {
	management.Elev.Floor = floor
}

func getFloor() int {
	return management.Elev.Floor
}

func setMotorStop() {
	elevio.SetMotorDirection(elevio.MotorDirStop)
}

func updateFloor(floor int) {
	if floor >= 0 {
		management.Elev.Floor = floor
		management.Elev.LastFloor = floor
		elevio.SetFloorIndicator(floor)
	}
}

func safeDrive() {

	orderManagement.UpdateCurrentOrder()
	orderManagement.UpdateMoveDir()

	if management.Elev.MoveDir == management.DirIdle {
		setMotorStop()
		return
	}

	setMotorFromDir()
	setElevState(nil, management.ElevMoving)
}

// -------------------------------------------------------------------------------------------
// State transitions
// -------------------------------------------------------------------------------------------
func setElevState(gs *orderManagement.GlobalState, state management.State) {
	prev := management.Elev.State
	management.Elev.State = state

	switch state {
	case management.ElevIdle:
		onIdleEntry(gs)
	case management.ElevMoving:
		onMovingEntry()
	case management.ElevStop:
		onStopEntry()
	case management.ElevObstruction:
		onObstructionEntry()
	}

	fmt.Println("STATE CHANGE:", prev, "->", state)
}

func onIdleEntry(gs *orderManagement.GlobalState) {
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	management.Elev.MoveDir = management.DirIdle
	safeDrive()
}

func onMovingEntry() {
	elevio.SetDoorOpenLamp(false)
	elevio.SetStopLamp(false)
	setMotorFromDir()
}

func onStopEntry() {
	elevio.SetStopLamp(true)
	stopElevator()
}

func onObstructionEntry() {
	stopElevator()
	elevio.SetDoorOpenLamp(true)
	if doorTimer != nil {
		doorTimer.Stop()
	}
	doorTimer = time.NewTimer(doorOpenDuration)
}
