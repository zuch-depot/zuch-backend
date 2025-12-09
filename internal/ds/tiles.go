package ds

type Tile struct {
	Tracks      [4]bool
	Signals     [4]bool
	IsPlattform bool
	IsBlocked   bool //nur für tracks
	ActiveTile  *ActiveTile
	X           int
	Y           int

	//ob auf dem Tile gebaut werden kann. Wenn es ein Hindernis ist, muss irgendwie bestimmt werden,
	//welches Sprite das ist, es können aber auch Tracks und Signale Mapseitig unveränderlich sein.
	IsLocked bool
}

type ActiveTile struct {
	Category   *ActiveTileCategory
	Name       string
	Level      int
	Stations   []*Station //Stationen, die in der Nähe sind. wird mit changeStationTile verwaltet
	Storage    map[string]int
	MaxStorage int //maximum Lager pro Gut -> sonst kann es zu unwiederruflichen auffüllen kommen
}

// Fügt bei i ein gleis hinzu, wenn da keins ist und das Tile nicht locked ist
// returnt true bei erfolg und false bei error
func (t *Tile) AddTrack(subtile int, gs *GameState) (bool, string) {
	if t.ActiveTile.Category != nil {
		return false, "There is a Active Tile there, no Tracks can be build on this tile."
	}
	if !t.Tracks[subtile-1] {
		t.Tracks[subtile-1] = true
		gs.BroadcastChannel <- WsEnvelope{Type: "rail.create", Username: "Server", Msg: &TileUpdateMSG{Position: [3]int{t.X, t.Y, subtile}}}
		return true, ""
	}
	return false, "There is already a Track at that Position."
}

// Entfernt bei i ein gleis und Signal, wenn da eins ist und das Tile nicht locked ist
// returnt true bei erfolg und false bei error
// wenn kein Signal an der Stelle ist, wird kein Fehler geworfen
func (t *Tile) RemoveTrack(subtile int, gs *GameState) (bool, string) {
	if t.Tracks[subtile-1] && !t.IsBlocked {
		t.Tracks[subtile-1] = false
		t.RemoveSignal(subtile, gs) // das muss noch kommunitzier werden
		gs.BroadcastChannel <- WsEnvelope{Type: "rail.remove", Username: "Server", Msg: &TileUpdateMSG{Position: [3]int{t.X, t.Y, subtile}}}
		return true, ""
	}
	return false, "There is no Track to Remove, or the Tile may be blocked by a Train, if so try again later."

}

// Fügt bei i ein Signal hinzu, wenn da keins ist und ein entsprechendes Gleis vorhanden ist,
// um bei i ein signal zu bauen muss gleis i da sein, und das Tile nicht locked ist;
// returnt true bei erfolg und false bei error
func (t *Tile) AddSignal(subtile int, gs *GameState) (bool, string) {
	if t.ActiveTile.Category != nil {
		return false, "There is a Active Tile there, no Signal can be build on this tile."
	}
	if t.Tracks[subtile-1] || t.Signals[subtile-1] {
		t.Signals[subtile-1] = true
		gs.BroadcastChannel <- WsEnvelope{Type: "signal.create", Username: "Server", Msg: &TileUpdateMSG{Position: [3]int{t.X, t.Y, subtile}}}
		return true, ""
	}
	return false, "There may be no Track to place the signal onto, or there is already a signal at that location."
}

// Fügt bei i ein Signal hinzu, wenn da keins ist und das Tile nicht locked ist
// returnt true bei erfolg und false bei error
func (t *Tile) RemoveSignal(subtile int, gs *GameState) (bool, string) {
	if t.Signals[subtile-1] {
		t.Signals[subtile-1] = false
		gs.BroadcastChannel <- WsEnvelope{Type: "signal.remove", Username: "Server", Msg: &TileUpdateMSG{Position: [3]int{t.X, t.Y, subtile}}}
		return true, ""
	}
	return false, "There is no Signal to remove."

}

// uhhh dunno was genau das hier ist
func (A *ActiveTile) calculateCargoPaths() {
}
