# Genereller Aufbau
- Jede Verbindung erfolgt über einen Websocket
- Der Server kann ohne aufforderung Nachrichten an den client senden und andersum
- Jede nachricht wird in einen **Envelope** verpackt
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
## Datenformate Bauwerke
### tileUpdateMSG
```golang
type tileUpdateMSG struct {
	Position [3]int
}
```
- die stellt nur eine Position dar, was dort passiert wird durch den typen bestimmt 
## Hinzufügen und entfernen von Zügen
### train.create
- nutzt die `trainCreateMSG` 
	- bei Typ kann angegeben werden was für ein waggong, dies bestimmt bspw. die kapazität. Muss mich da mit wilken aber noch absprechen
- antwortet mit einem Train
### train.remove
- nutzt die `trainRemoveMSG`
- antwortet mit einer `trainRemoveMSG`
## Datenformate hinzufügen und entfernen Züge
### trainCreateMSG
```go
type trainCreateMSG struct {
	Name string
	Waggons []trainCreateWaggons
	Id int
}
type trainCreateWaggons struct {
	Position [3]int
	Typ      string
}
```

### trainRemoveMSG
```go
type trainRemoveMSG struct {
	id int
}
```
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