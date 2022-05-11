// Package update provides methods for common update operations for SFO Museum Who's On First records.
// Methods signatures may still change.
package update

import (
	"context"
	"github.com/paulmach/orb/geojson"
	sfom_reader "github.com/sfomuseum/go-sfomuseum-reader"
	"github.com/whosonfirst/go-reader"
)

// Load a paulmach/orb/geojson.Feature for a Who's On First ID using a whosonfirst/go-reader.Reader instance.
func LoadFeature(ctx context.Context, r reader.Reader, id int64) (*geojson.Feature, error) {

	body, err := sfom_reader.LoadBytesFromID(ctx, r, id)

	if err != nil {
		return nil, err
	}

	f, err := geojson.UnmarshalFeature(body)

	if err != nil {
		return nil, err
	}

	return f, nil
}
