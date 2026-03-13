@echo off

start "Sim1" cmd /k "Simulator-v2-master\SimElevatorServer.exe --port 15657"
start "Sim2" cmd /k "Simulator-v2-master\SimElevatorServer.exe --port 15667"
REM start "Sim3" cmd /k "Simulator-v2-master\SimElevatorServer.exe --port 15677"
timeout /t 2

start "Elev1" cmd /k "go run . -simPort 15657 -peersPort 20001 -bcastPort 20002 -id 1"
start "Elev2" cmd /k "go run . -simPort 15667 -peersPort 20001 -bcastPort 20002 -id 2"
REM start "Elev3" cmd /k "go run . -simPort 15677 -peersPort 20001 -bcastPort 20002 -id 3"