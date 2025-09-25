package main

import "encoding/json"

type wsEnvelope struct {
	Type     string
	Username string
	Msg      any
}
type recieveWSEnvelope struct {
	Type     string
	Username string
	Msg      json.RawMessage
}

type tileUpdateMSG struct {
	X       int
	Y       int
	Subtile int // 1 => links, 2 => oben, 3 => rechts, 4 => unten
	// Wilken hat sich entschlossen immer wenn ein subtile als int gespeichert wird bei 1 anzufangen und wenn es ein bool[4] ist bei 0, also kann sein das es sich irgendwo verschiebt aber das kriegen wir sicher noch behoben Bei schienen auch analog
	Subject string //
	Action  string // remove, build
}

var (

	// #region Map & Tiles

	// map.initialLoad
	mapInitialLoad = wsEnvelope{Type: "game.initialLoad", Msg: &gamestate{}}
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
	trainMove = wsEnvelope{Type: "train.move", Msg: &Train{}}
	// #endregion Trains
// map.updateTile

)
