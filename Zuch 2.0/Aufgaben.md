Refactor main
- Abschaffen **createDemoTrains**
- **AddTrain** in ds verschieben und zu gamestate packen
- **checkIfWaggonsAreValid** ins ds verschieben
- **addTrain**
	- einmal ins ds und ggf. eine in kommunikation
- **calculateTrains**
	- in ds
- **processClientInputs** und alle Nachfolger in Kommunikation

Refactor ds
- **(*RecieveWSEnvelope).Reply** und **(*User).StartNotifiyingSingleClient** in kommunikation
- **(*Train).RemoveTrain** ggf. zu gamestate
- **AddStationTile** und **RemoveStationTile** ggf. zu gamestate
- **(*Station).RemoveStation** 
	- soll nicht changeStationTile aufrufen
	- zu gamestate schieben
- **(*Station).RemovePlattform**
	- Verbindungen zu changeStationTile rauslöschen
	- soll nur die Plattform entfernen
- Funktionen hinzufügen
	- AssignSchedule
	- UnAssignSchedule
- move() anpassen
	- nach jedem Signal neu berechnen

Kommunikation
- Methoden hinzufügen
	-  removeStop
	- addStop
	- changeSequenceSchedule
	- addWaggonToTrain
	- removeWaggonFromTrain
	- changeSequenceWaggons
	- Alle Namen anpassen können
		- Schedule
		- Stop wenn Wegpunkt
		- Stationen
		- Plattformen
		- Züge
	- Die Warenlogik, dass die Veränderungen und Lager angezeigt werden können



Frontend
- UIs
	- Zug erstellen/bearbeiten
	- Schedule erstellen/bearbeiten
	- overlay
- Map
	- anzeigen und bearbeiten
- Daten speichern
- Lobby?