Refactor main
- Abschaffen **createDemoTrains**
- **AddTrain** in ds verschieben und zu gamestate packen
	- fertig
- **checkIfWaggonsAreValid** ins ds verschieben
	- fertig
- **addTrain**
	- einmal ins ds und ggf. eine in kommunikation
	- ins ds gepackt, ggf. nochmal angucken
- **calculateTrains**
	- in ds
		- fertig
- **processClientInputs** und alle Nachfolger in Kommunikation

Refactor ds
- **(*RecieveWSEnvelope).Reply** und **(*User).StartNotifiyingSingleClient** in kommunikation
- **(*Train).RemoveTrain** ggf. zu gamestate
	- fertig
- **AddStationTile** und **RemoveStationTile** ggf. zu gamestate
	- fertig
- **(*Station).RemoveStation** 
	- soll nicht changeStationTile aufrufen
	- zu gamestate schieben
	- fertig
- **(*Station).RemovePlattform**
	- Verbindungen zu changeStationTile rauslöschen
	- soll nur die Plattform entfernen
	- entfernt die Plattform und alle Stops, die diese haben
	- fertig
- Funktionen hinzufügen
	- AssignSchedule
		- fertig
	- UnAssignSchedule
		- fertig
- move() anpassen
	- nach jedem Signal neu berechnen

Kommunikation
- siehe CLI Eingabeding Graph

# veraltet:
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