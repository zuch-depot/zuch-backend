package ds

import (
	"encoding/json"
	"log/slog"
)

//	es gibt beide, da ich json.RawMessage brauche um die nachricht zunächst nur zum teil dekodieren zu können. Da kann ich aber nicht alles reinschreiben also gibt es die normale variante die dafür ein any hat
//
// Mögliche Types siehe wsFormat.go, bspw. tile.update. Steuert welcher teil des Backends dafür zuständig ist
// bei Username wird automatisch der Username über den sich die websocket verbindung angemeldet hat eingetragen, Eingaben vom User werden überschrieben
// Msg ist die eigentliche nachricht, siehe ebenfalls wsFormat.go, nicht alle haben ein eigenes Struct, aber einige
type WsEnvelope struct {
	Type string
	Msg  any
}

func (envelope *RecieveWSEnvelope) Reply(success bool, message string, gs *GameState) {
	if envelope.TransactionID == "" {
		gs.Logger.Debug("Client did not transmit a TransactionID to track this Transaction, guess they dont want any feedback", slog.String("Username", envelope.User.Username))
		return
	}
	if !envelope.User.IsConnected {
		gs.Logger.Debug("Client no longer connected, guess they dont want any feedback", slog.String("Username", envelope.User.Username))
		return
	}

	replyInnerMSG := RelpyMSG{Success: success, Msg: message}
	reply := WsEnvelope{Type: "game.reply", Msg: replyInnerMSG}
	gs.Logger.Debug("Replying to Client event", slog.String("user", envelope.User.Username), slog.String("Event Type", reply.Type))
	envelope.User.WebSocketQueue <- reply

}

// beschreibung siehe typdefinition
func CreateGenericResponse(message string) *GenericResponse {
	return &GenericResponse{Body: struct {
		Message string
	}{Message: message}}
}

// ich wollte mir hier die optionen offen halten, vielleicht wird das ding nachher noch mit mehr infos angereichert aber das machen wa halt bisher noch nicht
// so ist es einfach nur ein string im body, vielleicht sollte man den noch zum Header machen, aber das kann man ja recht schnell nachher
type GenericResponse struct {
	Body struct {
		Message string
	}
}

type SaveGameResponse struct {
	Body struct {
		Message string
		Success bool
		Path    string
	}
}

//	es gibt beide, da ich json.RawMessage brauche um die nachricht zunächst nur zum teil dekodieren zu können. Da kann ich aber nicht alles reinschreiben also gibt es die normale variante die dafür ein any hat
//
// Mögliche Types siehe wsFormat.go, bspw. tile.update. Steuert welcher teil des Backends dafür zuständig ist
// bei Username wird automatisch der Username über den sich die websocket verbindung angemeldet hat eingetragen, Eingaben vom User werden überschrieben
// Msg ist die eigentliche nachricht, siehe ebenfalls wsFormat.go, nicht alle haben ein eigenes Struct, aber einige
type RecieveWSEnvelope struct {
	Type          string
	User          *User  // genutzt um zugriff auf die Connectin zu haben, damit man anhand der transactionID feedback / Fehler senden kann
	TransactionID string // Wird vom CLIENT gesetzt, der kann sich die dann merken und kriegt ggf, für fehler hier drüber eine rückmeldung vom Server
	Msg           json.RawMessage
}

// recht allgemeiner Typ der einfach 2 Positionen abbildet, wird für Tracks als auch bspw. Züge genutzt, die 2. position ist optional
type TileUpdateMSG struct {
	Position    *[3]int `path:"Position" example:"[1,3,4]" minLength:"3" maxLength:"3" doc:"Which Tile to interact with, in the format X Y Subtile, Subtile starts with 1 and on the left side, then counts clockwise"`
	Position_to *[3]int `path:"Position_to" example:"[1,1,4]" required:"false" minLength:"3" maxLength:"3" nullable:"true"`
	// 0 => links, 1 => oben, 2 => rechts, 3 => unten
	// Wilken hat sich entschlossen immer wenn ein subtile als int gespeichert wird bei 1 anzufangen und wenn es ein bool[4] ist bei 0, also kann sein das es sich irgendwo verschiebt aber das kriegen wir sicher noch behoben Bei schienen auch analog
}

type TrainMoveMSG struct {
	Id      int
	Waggons []*Waggon
}

type TrainCreateMSG struct {
	Name               string `example:"fred"`
	LocomotivePosition [3]int `example:"[1,1,4]" required:"true"`
	Id                 int    `hidden:"true"`
}

type TrainRemoveMSG struct {
	Id int
}

// blockiert mehrere, nicht nur ein tile
type BlockedTilesMSG struct {
	Tiles [][2]int
}

type TrainCreateWaggons struct {
	Position [3]int
	Typ      string // sollte mit einem string identifizieren wie schnell wie viel kapazität usw. muss mich da mit wilken noch genauer absprechen, bisher nur "Lebensmittel" als option
}

type RelpyMSG struct {
	Msg     string
	Success bool
}

type StationUpdateMsg struct {
	Position [2]int
}

type ScheduleCreateMSG struct {
	Name    string
	Entries []ScheduleEntry
}

type ScheduleRemoveMSG struct {
	Id int
}

type ScheduleEntry struct {
	PlattformId   int
	StationId     int
	LoadStrings   []string
	UnloadString  []string
	WaitTillFull  bool
	WaitTillEmpty bool
}

type ScheduleAssignMSG struct {
	ScheduleId int
	TrainId    int
}

type TileMessage struct {
	Body struct {
		Tiles [][]*Tile
	}
}
