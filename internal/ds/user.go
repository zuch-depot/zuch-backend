package ds

import (
	"log/slog"

	"github.com/gorilla/websocket"
)

type User struct {
	Username       string
	IsConnected    bool            `json:"-"`
	Connection     *websocket.Conn `json:"-"`
	WebSocketQueue chan WsEnvelope `json:"-"`
}

// bildet chatnahrichten ab
type ChatMessage struct {
	Nachricht string
	Sender    string
}

func (user *User) StartNotifiyingSingleClient(gs *GameState) {
	// Kombiniert den Broadcast und den einzelnen Channel, damit der client über die gleiche verbindung beides erhält
	for {
		if user.IsConnected {

			envelope := <-user.WebSocketQueue

			gs.Logger.Debug("Notifying client of event", slog.String("user", user.Username), slog.String("Event Type", envelope.Type))
			err := user.Connection.WriteJSON(envelope)
			if err != nil {
				gs.Logger.Error("Could not Write to client", slog.String("Error", err.Error()), slog.String("Username", user.Username))
			}
		} else {
			break
		}
	}
}
