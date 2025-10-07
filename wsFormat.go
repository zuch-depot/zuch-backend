package main

import (
	"encoding/json"
	"log/slog"
)

//	es gibt beide, da ich json.RawMessage brauche um die nachricht zunächst nur zum teil dekodieren zu können. Da kann ich aber nicht alles reinschreiben also gibt es die normale variante die dafür ein any hat
//
// Mögliche Types siehe wsFormat.go, bspw. tile.update. Steuert welcher teil des Backends dafür zuständig ist
// bei Username wird automatisch der Username über den sich die websocket verbindung angemeldet hat eingetragen, Eingaben vom User werden überschrieben
// Msg ist die eigentliche nachricht, siehe ebenfalls wsFormat.go, nicht alle haben ein eigenes Struct, aber einige
type wsEnvelope struct {
	Type          string
	Username      string
	TransactionID string
	Msg           any
}

//	es gibt beide, da ich json.RawMessage brauche um die nachricht zunächst nur zum teil dekodieren zu können. Da kann ich aber nicht alles reinschreiben also gibt es die normale variante die dafür ein any hat
//
// Mögliche Types siehe wsFormat.go, bspw. tile.update. Steuert welcher teil des Backends dafür zuständig ist
// bei Username wird automatisch der Username über den sich die websocket verbindung angemeldet hat eingetragen, Eingaben vom User werden überschrieben
// Msg ist die eigentliche nachricht, siehe ebenfalls wsFormat.go, nicht alle haben ein eigenes Struct, aber einige
type recieveWSEnvelope struct {
	Type          string
	user          *User  // genutzt um zugriff auf die Connectin zu haben, damit man anhand der transactionID feedback / Fehler senden kann
	TransactionID string // Wird vom CLIENT gesetzt, der kann sich die dann merken und kriegt ggf, für fehler hier drüber eine rückmeldung vom Server
	Msg           json.RawMessage
}

func (envelope *recieveWSEnvelope) reply(success bool, message string) {
	if envelope.TransactionID == "" {
		logger.Debug("Client did not transmit a TransactionID to track this Transaction, guess they dont want any feedback", slog.String("Username", envelope.user.username))
		return
	}
	if !envelope.user.isConnected {
		logger.Debug("Client no longer connected, guess they dont want any feedback", slog.String("Username", envelope.user.username))
		return
	}

	replyInnerMSG := relpyMSG{Success: success, Msg: message}
	reply := wsEnvelope{Type: "game.reply", Username: envelope.user.username, TransactionID: envelope.TransactionID, Msg: replyInnerMSG}
	logger.Debug("Replying to Client event", slog.String("user", envelope.user.username), slog.String("Event Type", reply.Type))
	envelope.user.webSocketQueue <- reply

}

type tileUpdateMSG struct {
	position [3]int // 0 => links, 1 => oben, 2 => rechts, 3 => unten
	// Wilken hat sich entschlossen immer wenn ein subtile als int gespeichert wird bei 1 anzufangen und wenn es ein bool[4] ist bei 0, also kann sein das es sich irgendwo verschiebt aber das kriegen wir sicher noch behoben Bei schienen auch analog
	Subject string //
	Action  string // remove, build
}

type trainMoveMSG struct {
	Id      int
	Waggons []*Waggon
}

type trainCreateMSG struct {
	Name    string
	Waggons []trainCreateWaggons
	Id      int
}

type trainCreateWaggons struct {
	Position [3]int
	Typ      string
}

type relpyMSG struct {
	Msg     string
	Success bool
}

var (

	// #region game
	// map.initialLoad
	// Wird genutzt um anfangs den aktuellen stand an den client zu senden, hier ist so gut wie alles enthalten das das backend weiß
	// Strukturiert wie das gamestate objekt, nutzt game.initialLoad als type
	mapInitialLoad = wsEnvelope{Type: "game.initialLoad", Msg: &gamestate{}}
	gameReplyMsg   = wsEnvelope{Type: "game.reply", Msg: &relpyMSG{}}
	// #endregion game

	// #region Map & Tiles
	// Aktualisierung von genau einem Tile, bspw. Schiene oder Signal bauen
	// 	{
	//   "Type": "tile.update",
	//   "Msg": {
	//     "X": 0,
	//     "Y": 0,
	//     "Subtile": 3,  // geht von 1-4
	//     "Subject": "rail",
	//     "Action": "build"
	//   }
	//  }
	mapUpdate = wsEnvelope{Type: "tile.update", Msg: &tileUpdateMSG{}}

	// #endregion Map

	// #region Trains
	// wird bisher genutzt um die bewegung von zügen darzustellen
	// Bezieht sich auch auf genau einen zug
	trainMove = wsEnvelope{Type: "train.move", Msg: &trainMoveMSG{}}
	// Eingehend
	trainCreateIn = wsEnvelope{Type: "train.create", Msg: &trainCreateMSG{}}
	// Ausgehend
	trainCreateOut = wsEnvelope{Type: "train.create", Msg: &Train{}}

	// #endregion Trains
// map.updateTile

)
