VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGTVERALTET NEUE DOKU FOLGT
VERALTET NEUE DOKU FOLGT

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
- **ggf. erhält der Client eine game.reply gefolgt von einer generellen Nachricht an alle clients die die eigentliche änderung übermittelt. game.reply wird nur genutzt um fehler anzeigen zu können. Änderungen werden immer anderwaltig übermittelt**
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

```go
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

## Hinzufügen, entfernen und bewegen von Zügen

### train.create

- nutzt die `trainCreateMSG`
  - bei Typ kann angegeben werden was für ein waggong, dies bestimmt bspw. die kapazität. Muss mich da mit wilken aber noch absprechen
- antwortet mit einem Train
- wird als still stehender zug erstellt. Hat allerdings schon alle seine waggons, ggf kann man da später noch welche hinzufügen

### train.remove

- nutzt die `trainRemoveMSG`
- antwortet mit einer `trainRemoveMSG`

### train.move

- nutzt die trainMoveMSG
- wird nur vom server ausgehend gesendet

### train.cargochange

- wird ausgesendet wenn der Füllstand eines zuges sich ändert
- schickt einfach einen neuen `zug`, langsam wirds mir zu bunt
- wird nur vom server aus getriggert, client kann die nicht schicken

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
type TrainMoveMSG struct {
	Id      int
	Waggons []*Waggon
}
```

```json
{
  "Type": "train.create",
  "TransactionID": "edabad9e-e1a7-4e91-977a-119daaa8775e",
  "Msg": {
    "Name": "hallo",
    "Waggons": [
      {
        "Position": [8, 6, 2],
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

### ZUCHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHH choo choo

```go
// gesetzt werden müssen: Name, Waggons, Id
type Train struct {
	Name               string    // Nicht eindeutig, dafür siehe ID
	Waggons            []*Waggon //Alle müssen nebeneinander spawnen
	Schedule           *Schedule
	NextStop           Stop     //der Stop, in dem der Zug gerade hinfährt oder plant hinzufahren. Wenn 0, dann nächster Stop aus Schedule
	CurrentPath        [][3]int //neu berechnen bei laden
	CurrentPathSignals [][3]int

	Waiting         bool //hat letzten Tick ein geblockes Tile gefunden oder keinen Weg gefunden und wartet
	LoadingTime     int  //Wie lange ist der Zug schon am be-/entladen? 0 == nicht am laden. Zeiteinheit ist wie oft methode aufgerufen wurde
	FinishedLoading bool //wenn nichts mehr geladen wird true. Kann auch wieder zurückgenommen werden
	Id              int
	//User            *User nur ggf., falls wird das implementieren
}

type Waggon struct {
	Position     [3]int //x,y,sub
	MaxSpeed     int
	CargoStorage *CargoStorage
}

type CargoStorage struct {
	Capacity        int //statisch
	Filled          int
	FilledCargoType string
	CargoCategory   string //statisch
}
```

## Blockieren und entblocken von Tiles

- wenn züge sich einen weg reservieren wird dieser blockiert
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
{
  // hier ist halt der ganze envelope, nicht nur die MSG wie im go beispiel
  "Type": "tiles.unblock",
  "Username": "Server",
  "TransactionID": "",
  "Msg": {
    "Tiles": [[1, 3]]
  }
}
```

## Erstellen und zuordnen von Schedules

- Eine schedule besteht aus mehreren Stops, bei den Stops können die züge jeweils eineen Load oder Unload command haben und nehmen dementsprechend dort ware auf oder geben sie ab

### schedule.create

- erstellt eine neue schedule mit den gegebenen stops und
- nutzt die `ScheduleCreateMsg` zum erstellen
- antwortet mit der neuen `Schedule`

### schedule.remove

- löscht eine schedule
- Empfängt und sendet eine `ScheduleRemoveMSG`

### schedule.assign

- weißt einem zug eine schedule zu
- empfängt eine `ScheduleAssignMSG`, Schickt einen neuen zug zurück

### schedule.unassign

- entfernt eine schedule vom einem zug
- nichts von beidem wird dadurch gelöscht
- empfängt eine `ScheduleAssignMSG`, Schickt einen neuen zug zurück

## Datenformate erstellen und zuordnen von schedules

### ScheduleAssignMSG

```go
type ScheduleAssignMSG struct {
	ScheduleId  int
	TrainId 	int
}
```

### ScheduleCreateMSG

```go
type ScheduleCreateMSG struct {
	Name    string
	Entries []ScheduleEntry
}

type ScheduleEntry struct {
	PlattformId   int
	StationId     int
	LoadStrings   []string // gibt an welche güter geladen werden, bspw. "Lebensmittel"
	WaitTillFull  bool
	UnloadString  []string
	WaitTillEmpty bool
}
```

### ScheduleRemoveMSG

```go
type ScheduleRemoveMSG struct {
	Id int
}
```
