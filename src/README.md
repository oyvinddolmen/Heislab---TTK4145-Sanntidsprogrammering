# Elevator Project
## Project Description
We were tasked with creating software for controlling n elevators working in parallel across m floors. Details about the implementation of the solution were not specified, but the system had to fulfill some main requirements, as well as a secondary requirement regarding system efficiency.
**Main requirements:**
  - The button lights are a service guarantee.
  - No calls are lost.
  - The lights and buttons should function as expected.
  - The door should function as expected.
  - An individual elevator should behave sensibly and efficiently.

**Secondary requirements:**
  - Calls should be served as efficiently as possible.


## Our system
Our system is designed based on a peer-to-peer solution. 

Elevators communicate through UDP broadcasts containing each elevators view of the global state. Upon changes in an elevators view of the global state, a broadcast is sent to the other elevators. In addition, each elevator periodically broadcasts their view. This periodic broadcast works as a heartbeat, with each elevator tracking the time since the last heartbeat of its peers, and setting them to offline if they take too long. 

Our system is mostly implemented in Go.

### Modules
#### Main


#### State


#### Network


#### Elevator


#### ElevIO ?


#### OrderManagement


#### Management