# Heislab---TTK4145-Sanntidsprogrammering
Bombesikkert system for tre heiser

Hvordan jobbe i branches og merge:

gå til ny branch:
    git checkout Tobias

hente main:
    git rebase main

    hvis konflikt: 
        løs konfliktene i filene
        git add . 
        git rebase --continue

gjør arbeid .... deretter legge til endringene i main:
    git add .
    git commit -m meldingen din

bytt over til main og pull inn endringene fra branchen din
    git checkout main
    git pull
    git merge Tobias
    git push


test
