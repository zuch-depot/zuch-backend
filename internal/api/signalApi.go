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
	},
	) (*ds.GenericResponse, error) {
		var cost int
		tile, err := gs.GetTile(i.Body.Position[0], i.Body.Position[1])
		if err != nil {
			return nil, fmt.Errorf("Could not find Tile; %s", err.Error())
		}
		cost, err = tile.AddSignal(i.Body.Position[2], gs, !i.CostsOnly)
		if err != nil {
			return nil, fmt.Errorf("Tile was found but could not create signal; %s", err.Error())
		}
		return ds.CreateGenericResponse("created signal", cost), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"signal"}
		o.Summary = "Signal bauen"
	})

	huma.Delete(*api, "/signal", func(ctx context.Context, i *struct {
		CostsQuery
		Body ds.TileUpdateMSG
	},
	) (*ds.GenericResponse, error) {
		var refund int
		tile, err := gs.GetTile(i.Body.Position[0], i.Body.Position[1])
		if err != nil {
			return nil, fmt.Errorf("Tile not found; %s", err.Error())
		}
		refund, err = tile.RemoveSignal(i.Body.Position[1], gs, !i.CostsOnly)
		if err != nil {
			return nil, fmt.Errorf("Tile was found but could not remove signal; %s", err.Error())
		}
		return ds.CreateGenericResponse("removed signal", -refund), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"signal"}
		o.Summary = "Signal zerstören"
	})
}

// endregion signals
