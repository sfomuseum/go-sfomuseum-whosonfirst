package custom

import (
	"context"
	"fmt"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v2"
	"io"
	"strings"
)

// SFOMuseumPropertiesOptions is a struct containing configuration option for updating Who's On First records with SFO Museum specific properties.
type SFOMuseumPropertiesOptions struct {
	// DataReader is a `whosonfirst/go-reader.Reader` instance used to read Who's On First records.
	DataReader reader.Reader
	// DataReader is a `whosonfirst/go-writer.Writer` instance used to write Who's On First records.
	DataWriter writer.Writer
	// PropertiesReader is a a `whosonfirst/go-reader.Reader` instance used to read SFO Museum properties.
	PropertiesReader reader.Reader
	// PropertiesWriter is a a `whosonfirst/go-writer.ReaderWriter` instance used to writer SFO Museum properties.
	PropertiesWriter writer.Writer
	// CustomProperties is a dictionary containing SFO Museum properties to append to a Who's On First record.
	CustomProperties map[string]interface{}
}

// ApplySFOMuseumProperties updates one or more Who's On First records identified by 'ids' with SFO Museum specific properties.
func ApplySFOMuseumProperties(ctx context.Context, opts *SFOMuseumPropertiesOptions, ids ...int64) error {

	// TBD: Do this concurrently?

	for _, id := range ids {

		err := applySFOMuseumProperties(ctx, opts, id)

		if err != nil {
			return fmt.Errorf("Failed to apply sfomuseum properties for %d, %w", id, err)
		}
	}

	return nil
}

func applySFOMuseumProperties(ctx context.Context, opts *SFOMuseumPropertiesOptions, id int64) error {

	data_path, err := uri.Id2RelPath(id)

	if err != nil {
		return fmt.Errorf("Failed to derive path for %d, %w", id, err)
	}

	data_fh, err := opts.DataReader.Read(ctx, data_path)

	if err != nil {
		return fmt.Errorf("Failed read data for %s, %w", data_path, err)
	}

	data_body, err := io.ReadAll(data_fh)

	if err != nil {
		return fmt.Errorf("Failed to read feature body, %w", err)
	}

	data_pt, err := properties.Placetype(data_body)

	if err != nil {
		return fmt.Errorf("Failed to derive placetype for %d, %w", id, err)
	}

	props, err := EnsureCustomProperties(ctx, opts.PropertiesReader, opts.PropertiesWriter, id)

	if err != nil {
		return fmt.Errorf("Failed to read custom properties for %d, %w", id, err)
	}

	props["wof:repo"] = "sfomuseum-data-whosonfirst"

	switch data_pt {
	case "campus":
		props["sfomuseum:placetype"] = "airport"
	case "locality":
		props["sfomuseum:placetype"] = "city"
	default:
		props["sfomuseum:placetype"] = data_pt
	}

	if opts.CustomProperties != nil {

		for k, v := range opts.CustomProperties {
			props[k] = v
		}

	}

	props = ApplyEDTFFixes(ctx, data_body, props)

	err = WriteCustomProperties(ctx, opts.PropertiesWriter, id, props)

	if err != nil {
		return fmt.Errorf("Failed to write custom properties for %d, %w", id, err)
	}

	/*

	> go build -o bin/import-feature cmd/import-feature/main.go && ./bin/import-feature -access-token-uri 'file:///usr/local/sfomuseum/lockedbox/github/geotag?decoder=string' 85951061
	2022/08/08 17:23:48 processing (85951061)
	2022/08/08 17:23:49 Add sfomuseum-data-whosonfirst/data/859/510/61/85951061.geojson @data-27888
	2022/08/08 17:23:49 processing (85633793)
	2022/08/08 17:23:49 processing (85688599)
	2022/08/08 17:23:50 Add sfomuseum-data-whosonfirst/data/856/885/99/85688599.geojson @data-27888
	2022/08/08 17:23:51 Add sfomuseum-data-whosonfirst/data/856/337/93/85633793.geojson @data-27888
	2022/08/08 17:24:03 Add sfomuseum-data-whosonfirst/properties/859/510/61/85951061.json @props-27888
	2022/08/08 17:24:05 Add sfomuseum-data-whosonfirst/properties/859/510/61/85951061.json @props-27888
	2022/08/08 17:24:07 Failed to import IDs, Failed to import IDs, Failed to apply SFO Museum properties, Failed to apply sfomuseum properties for 85951061, Failed to merge custom properties for 85951061, Failed to read custom properties for 85951061, Failed to read custom properties for 859/510/61/85951061.json, Unexpected status: 404 Not Found

	*/

	err = MergeCustomProperties(ctx, opts.PropertiesReader, opts.DataReader, opts.DataWriter, id)

	if err != nil && strings.HasSuffix(err.Error(), "404 Not Found") {
		//
	}
	
	if err != nil {
		return fmt.Errorf("Failed to merge custom properties for %d, %w", id, err)
	}

	return nil
}
