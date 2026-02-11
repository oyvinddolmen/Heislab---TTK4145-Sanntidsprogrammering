# Heislab---TTK4145-Sanntidsprogrammering
Bombesikkert system for tre heiser

for å starte på opp den fysiske heisen skriv: 
    "elevatorserver" i terminalen

for å starte opp simulatoren: 
    In terminal inside Simulator folder: 
    dmd -w -g src\sim_server.d src\timer_event.d -ofSimElevatorServer.exe
    .\SimElevatorServer.exe
 


Hvordan jobbe i branches og merge:

VIKtIG: start med å hente nye endringer i main
- git pull main

gå til ny branch:
- git checkout Tobias

hente main:
- git rebase main

hvis konflikt: 
- løs konfliktene i filene
- git add .
- git rebase --continue

gjør arbeid .... deretter legge til endringene i main:
- git add .
- git commit -m meldingen din

bytt over til main og pull inn endringene fra branchen din
- git checkout main
- git pull
- git merge Tobias
- git push
