# Genereller Aufbau
- Jede Verbindung erfolgt über einen Websocket
- Der Server kann ohne aufforderung Nachrichten an den client senden und andersum
- Jede nachricht wird in einen **Envelope** verpackt
- im wsbeispiele ordner sind beispielanfragen, können mit 
## Funktionsweise
- der client verbindet sich per Websocket zum Backend
- dort wird automatisch ein "Spieler" für ihn erstellt
- von nun an kommuniziert der Server events über den Websocket an den client, bspw. wenn ein anderer Spieler einen zug gebaut hat
- Der client kann nun ebenfalls über den selben Websocket anfragen an den Server schicken, dies geschieht über den Envelope
- Der Server antwortet und sendet events über den gleichen websocket
- Der Server schickt aber nicht immer alles an alle spieler, bspw. Antworten auf anfragen werden nur an den auslöser gesendet, diese sind mit **game.reply** gekennzeichnet
- **ggf. erhält der Client eine game.reply gefolgt  von einer generellen Nachricht an alle clients  die die eigentliche änderung übermittelt. game.reply wird nur genutzt um fehler anzeigen zu können. Änderungen werden immer anderwaltig übermittelt**
- Hier sind die GO datentypen dargestellt. Versendet wird das ganze aber als JSON, die namen der felder bleiben aber erhalten also kann auch in Typescript mit envelope.Type auf bspw. den typen zugegriffen werden, wenn ich mal ganz viel langeweile hab kommen hier vielleicht auch noch Json beispiele dazu 
## Der Envelope
- Intern werden hier 2 leicht verschiedene arten an Envelopes verwendet, Sind aber beide equivalent wenn man von außen mit ihnen hantiert
```golang
type wsEnvelope struct {
	Type          string
	TransactionID string
	Msg           any
}
```
### Type
- hier wird angegeben was man eigentlich machen will
- hier könnte beispielsweise "train.create" stehen
- das sorgt dann dafür das es intern richtig ausgewertet wird und muss somit gesetzt sein, 
- Eine ausführliche Liste aller möglichen anfragen und antworten folgt 
## TransactionID
- Der Server muss ggf. auf Nachrichten antworten können
- also kann der **CLIENT** eine Transaction Id mitgeben damit er nachher weiß auf was der server eigentlich antwortet
- gibt der client keine Transaction Id mit **antwortet der Server ihm NICHT auf einzelne anfragen**, da ohnehin die antwort auf keine anfrage zurückgeführt werden kann
## Msg
- Hier kommen die **Parameter** entsprechend des Type ein
- bspw. kann im Type angegeben werden "train.create", dann müsste in der Msg ein object mit den passenden Feldern um einen zug zur welt zu bringen sein. Infos dazu siehe unten

# Mögliche Nachrichten
- Die überschriften stellen die typen der anfrage dar
- dann folgen details zu anfrage und antwort
## Initialer Stand
### game.initialLoad
- game.initialLoad übermittelt zunächst den aktuellen stand des Spiels. Dieser soll vom client dann kontinuirlich aktualisiert werden. 
- Nutzt als daten Struktur den `SendAbleGamestate`
## Datenformate Initlialer Stand
### SendableGameState
``` go
type SendAbleGamestate struct {
	Users     []*User
	Schedules []*Schedule
	Stations  map[int]*Station
	Tiles     [][]*Tile
	Trains    map[int]*Train
}
``` 

## erstellen und löschen von Bauwerken (signale Schienen und später Bahnhöfe)
### Signale
#### signal.create
- anfragen und antworten per `tileUpdateMSG`
#### signal.remove
- anfragen und antworten per `tileUpdateMSG`
### Gleise
#### rail.create
- anfragen und antworten per `tileUpdateMSG`
#### rail.remove
- anfragen und antworten per `tileUpdateMSG`
### Stationen
#### station.create
- anfragen per `tileUpdateMSG`
- Als Antwort wird erstmal die ganze neue `Station` gesendet, da die etwas komplexer sind
- Stationen können nur gebaut werden wenn da bereits gleise liegen
#### station.remove
- anfragen per `tileUpdateMSG`
- Als Antwort wird erstmal die ganze neue `Station` gesendet, da die etwas komplexer sind, also nicht nur der teil der jettz gelöscht ist, sondern einmal alles was jetzt noch da ist, der aktuelle stand
- **entfernt nur ein tile der station, um die station komplett zu löschen muss jedes tile einzeln entfernt werden**
- **Beim entfernen des letzten Tiles läuft es noch nicht so run**

## Datenformate Bauwerke
### tileUpdateMSG
```golang
type tileUpdateMSG struct {
	Position [3]int
}
```
- die stellt nur eine Position dar, was dort passiert wird durch den typen bestimmt 
## Hinzufügen, entfernen und setzen von Schedules bei Zügen
### train.create
- nutzt die `trainCreateMSG` 
	- bei Typ kann angegeben werden was für ein waggong, dies bestimmt bspw. die kapazität. Muss mich da mit wilken aber noch absprechen
- antwortet mit einem Train
- wird als still stehender zug erstellt. Hat allerdings schon alle seine waggons, ggf kann man da später noch welche hinzufügen
### train.remove
- nutzt die `trainRemoveMSG`
- antwortet mit einer `trainRemoveMSG`
### train.assignSchedule
## Datenformate hinzufügen und entfernen Züge
### trainCreateMSG
```go
type trainCreateMSG struct {
	Name	string
	Waggons []trainCreateWaggons
	Id		int
}
type trainCreateWaggons struct {
	Position [3]int
	Typ      string
}

```
```json

{
  "Type": "train.create",
  "TransactionID":"edabad9e-e1a7-4e91-977a-119daaa8775e",
  "Msg": {
    "Name": "hallo",
    "Waggons": [
      {
        "Position": [8,6,2],
        "Typ": "Lebensmittel"
      }
    ]
  }
}
```

### trainRemoveMSG
```go
type trainRemoveMSG struct {
	id int
}
```

<!-- ## Erstellen und löschen von Schedules
### schedule.create
- genutzt um basierend auf einer liste an stops eine schedule zu erstellen
- es erfolgt keine prüfung ob die wirklich fahrbar ist

### schedule.remove
## Datenformate Erstellen und löschen von Schedules
## Schedule.create
```go
type scheduleCreateMsg struct {
	Name 	string
	Stops	[]Stop
}
``` -->
## Blockieren und entblocken von Tiles 
- wenn  züge sich einen weg reservieren wird dieser blockiert
- wäre nett das irgendwie sehen zu können
- **geht immer vom Server aus** und nie vom client
### tiles.block
- sendet eine blockedTilesMSG, wird nur vom server verwendet. client sendet die eh nicht
### tiles.unblock
- sendet eine blockedTilesMSG, wird nur vom server verwendet. client sendet die eh nicht
## Datenformate blockieren und entblockieren
```go
type blockedTilesMSG struct {
	Tiles [][2]int
}
```
```json
{ // hier ist halt der ganze envelope, nicht nur die MSG wie im go beispiel
  "Type": "tiles.unblock",
  "Username": "Server",
  "TransactionID": "",
  "Msg": {
    "Tiles": [
      [
        1,
        3
      ]
    ]
  }
}
```
## Erstellen und zuordnen von Schedules
- Eine schedule besteht aus mehreren Stops, bei den Stops können die züge jeweils eineen Load oder Unload command haben und nehmen dementsprechend dort ware auf oder geben sie ab 
### schedule.create 
- erstellt eine neue schedule mit den gegebenen stops und 