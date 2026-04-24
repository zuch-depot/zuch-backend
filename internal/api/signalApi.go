package api

import (
	"context"
	"fmt"
	"zuch-backend/internal/ds"

	"github.com/danielgtaylor/huma/v2"
)

// region Signals
// Hier werden die HUMA Routen für die Signale erstellt
// Heißt hier werden auch die eigentlichen methoden aufgerufen und fehler abgefangen
func registerSignalRoutes(api *huma.API, gs *ds.GameState) {
	// hier kann man ein oder mehr signale bauen
	huma.Post(*api, "/signal", func(ctx context.Context, i *struct {
		CostsQuery
		Body ds.TileUpdateMSG
	}) (*ds.GenericResponse, error) {
		tile, err := gs.GetTile(i.Body.Position[0], i.Body.Position[1])
		if err != nil {
			return nil, fmt.Errorf("Could not find Tile; %s", err.Error())
		}
		err = tile.AddSignal(i.Body.Position[2], gs)
		if err != nil {
			return nil, fmt.Errorf("Tile was found but could not create signal; %s", err.Error())
		}
		return ds.CreateGenericResponse("created signal"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"signal"}
		o.Summary = "Signal bauen"
	})

	huma.Delete(*api, "/signal", func(ctx context.Context, i *struct {
		CostsQuery
		Body ds.TileUpdateMSG
	}) (*ds.GenericResponse, error) {
		tile, err := gs.GetTile(i.Body.Position[0], i.Body.Position[1])
		if err != nil {
			return nil, fmt.Errorf("Tile not found; %s", err.Error())
		}
		err = tile.RemoveSignal(i.Body.Position[2], gs)
		if err != nil {
			return nil, fmt.Errorf("Tile was found but could not remove signal; %s", err.Error())
		}
		return ds.CreateGenericResponse("removed signal"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"signal"}
		o.Summary = "Signal zerstören"
	})
}

// endregion signals
