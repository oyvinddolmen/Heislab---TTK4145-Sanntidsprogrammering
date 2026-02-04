# Heislab---TTK4145-Sanntidsprogrammering
Bombesikkert system for tre heiser

Hvordan jobbe i branches og merge:

For å være up to date med main
    git rebase main

lage branch: 
    git checkout -b feature/network

commite og pushe: 
    git status
    git add .
    git commit -m "Implement UDP heartbeat"
    git push -u origin feature/network

