package api

import (
	"context"
	"fmt"

	"zuch-backend/internal/ds"

	"github.com/danielgtaylor/huma/v2"
)

func registerActiveTileRoutes(api *huma.API, gs *ds.GameState) {
	// hier kann man alle auflisten
	huma.Get(*api, "/activetiles", func(ctx context.Context, i *struct{}) (*struct {
		Body struct {
			ActiveTiles map[int]*ds.ActiveTile
		}
	}, error,
	) {
		return &struct {
			Body struct{ ActiveTiles map[int]*ds.ActiveTile }
		}{
			Body: struct{ ActiveTiles map[int]*ds.ActiveTile }{
				ActiveTiles: gs.ActiveTiles,
			},
		}, nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"ActiveTiles"}
		o.Summary = "ActiveTiles auflisten"
	})

	// hier kann man eine active Tile umbennen
	huma.Post(*api, "/activetiles/{id}", func(ctx context.Context, i *struct {
		Id   int `path:"id" example:"1"`
		Body struct {
			Name string `example:"fred"`
		}
	},
	) (*ds.GenericResponse, error) {
		atile, ok := gs.ActiveTiles[i.Id]
		if !ok {
			return nil, fmt.Errorf("active Tile does not exist")
		}
		err := atile.Rename(i.Body.Name, gs)
		if err != nil {
			return nil, fmt.Errorf("could not rename Tile ; %s", err.Error())
		}
		return ds.CreateGenericResponse("renamed Tile"), nil
	}, func(o *huma.Operation) {
		o.Tags = []string{"ActiveTiles"}
		o.Summary = "ActiveTile umbennen"
	})
}
