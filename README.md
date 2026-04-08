# Heislab---TTK4145-Sanntidsprogrammering
    
Project Description:

We were tasked with creating software for controlling n elevators working in parallel across m floors. Details about the implementation of the solution were not specified, but the system had to fulfill some main requirements, as well as a secondary requirement regarding system efficiency. Main requirements:

    The button lights are a service guarantee.
    No calls are lost.
    The lights and buttons should function as expected.
    The door should function as expected.
    An individual elevator should behave sensibly and efficiently.

Secondary requirements:

    Calls should be served as efficiently as possible.

Our system

In our implementation each elevator runs as an independent peer with its own local finite state machine, while also sharing a common view of the global system state over the network.

The solution is based on peer-to-peer communication using UDP broadcast. Elevators continuously exchange their current state and hall requests, allowing the system to coordinate order distribution without a central controller. To handle failures, each elevator monitors heartbeats from the others and marks missing peers as offline if they stop responding.

Hall orders are assigned to the most suitable active elevator, while cab orders are handled locally by each elevator. This gives a fault-tolerant and efficient system where no requests are lost, button lights act as a service guarantee, and the elevators continue to operate sensibly even if one or more peers disconnects.


For å starte opp simulatoren: 
    In terminal inside Simulator folder: 
    dmd -w -g src\sim_server.d src\timer_event.d -ofSimElevatorServer.exe
    .\SimElevatorServer.exe --port 15657
    .\SimElevatorServer.exe --port 15667

Starte programmet:
    go run . -simPort 15657 -bcastPort 20002 -id 1
    go run . -simPort 15667 -bcastPort 20002 -id 2




