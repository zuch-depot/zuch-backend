package ds

var (

	// #region game
	// map.initialLoad
	// Wird genutzt um anfangs den aktuellen stand an den client zu senden, hier ist so gut wie alles enthalten das das backend weiß
	// Strukturiert wie das gamestate objekt, nutzt game.initialLoad als type
	mapInitialLoad = WsEnvelope{Type: "game.initialLoad", Msg: &SendAbleGamestate{}}

	// Wird genutzt um antworten auf die "Anfragen" des Client zu schicken, geht immer nur an den client der sie geschickt hat, der client kann bei seinen anfragen eine Transaktion ID eintragen, die wird hier auch wieder rein kopiert
	gameReplyMsg = WsEnvelope{Type: "game.reply", Msg: &RelpyMSG{}}
	// #endregion game

	// #region Map & Tiles
	railCreate   = WsEnvelope{Type: "rail.create", Msg: &TileUpdateMSG{}}
	railRemove   = WsEnvelope{Type: "rail.remove", Msg: &TileUpdateMSG{}}
	signalCreate = WsEnvelope{Type: "signal.create", Msg: &TileUpdateMSG{}}
	signalRemove = WsEnvelope{Type: "signal.remove", Msg: &TileUpdateMSG{}}

	tilesBlock   = WsEnvelope{Type: "tiles.block", Msg: &BlockedTilesMSG{}}
	tilesUnblock = WsEnvelope{Type: "tiles.unblock", Msg: &BlockedTilesMSG{}}

	// #endregion Map & Tiles

	// #region Trains
	// wird bisher genutzt um die bewegung von zügen darzustellen
	// Bezieht sich auch auf genau einen zug
	trainMove = WsEnvelope{Type: "train.move", Msg: &TrainMoveMSG{}}
	// Eingehend
	trainCreateIn = WsEnvelope{Type: "train.create", Msg: &TrainCreateMSG{}}
	trainRemove   = WsEnvelope{Type: "train.remove", Msg: &TrainRemoveMSG{}}
	// Ausgehend
	trainCreateOut = WsEnvelope{Type: "train.create", Msg: &Train{}}

	// #endregion Trains

// map.updateTile

)
