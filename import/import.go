package whosonfirst

import (
	"context"
	"fmt"
	"github.com/sfomuseum/go-sfomuseum-whosonfirst/custom"
	"github.com/whosonfirst/go-whosonfirst-fetch"
)

type ImportFeatureOptions struct {
	Fetcher           *fetch.Fetcher
	BelongsTo         []string
	PropertiesOptions *custom.SFOMuseumPropertiesOptions
}

func ImportFeatures(ctx context.Context, opts *ImportFeatureOptions, ids ...int64) error {

	for _, id := range ids {

		fetched_ids, err := opts.Fetcher.FetchIDs(ctx, []int64{id}, opts.BelongsTo...)

		if err != nil {
			return fmt.Errorf("Failed to fetch IDs, %v", err)
		}

		err = custom.ApplySFOMuseumProperties(ctx, opts.PropertiesOptions, fetched_ids...)

		if err != nil {
			return fmt.Errorf("Failed to apply SFO Museum properties, %v", err)
		}
	}

	return nil
}
