package api

import (
	"context"
	"zuch-backend/internal/ds"

	"github.com/danielgtaylor/huma/v2"
)

// region tiles
func registerTileRoutes(api *huma.API, gs *ds.GameState) {
	huma.Get(*api, "/tiles", func(ctx context.Context, i *struct{}) (*ds.TileMessage, error) {
		mes := &ds.TileMessage{}
		mes.Body.Tiles = gs.Tiles
		return mes, nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"tiles"}
		o.Summary = "Tiles auflisten"
	})
}

// endregion tiles
