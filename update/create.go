package update

import (
	"context"
	"fmt"
	"github.com/paulmach/orb/geojson"
	"github.com/sfomuseum/go-edtf"
	sfom_reader "github.com/sfomuseum/go-sfomuseum-reader"
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-writer"
	"strings"
)

// Create a new Who's On First record and write it to 'wr'. The signature for this method is a bit of a mess right now
// and reflects the specific concerns of the go-sfomuseum-gis/cmd/sfo-export-complex tool where it was first written.
// 'tmp_f' is assumed to be a GeoJSON records that was exported from the SFO GIS system so we are mostly only concerned
// with its geometry and a handful of properties. Which is another way of saying 'tmp_f' is assumed to contain a bunch
// of ESRI-specific properties we don't want to export. 'parent_id' is passed in explicitly in order to append its hierarchy
// using 'r' to load the parent feature. 'extra_props' are things that used to set explicitly or passed in as atomic values
// namely wof:repo and edtf:inception.
func CreateWithFeature(ctx context.Context, r reader.Reader, wr writer.Writer, tmp_f *geojson.Feature, parent_id int64, extra_props map[string]interface{}) (int64, error) {

	new_f := geojson.NewFeature(tmp_f.Geometry)
	new_props := new_f.Properties

	tmp_props := tmp_f.Properties

	for key, value := range tmp_props {

		if strings.HasPrefix(key, "sfomuseum:") {

			if key == "sfomuseum:key" {
				continue
			}

			new_props[key] = value
			continue
		}

		switch key {
		case "sfo:id", "sfo:level":
			new_props[key] = value
		default:
			// pass
		}
	}

	// really?
	for k, v := range extra_props {
		new_props[k] = v
	}

	ensure := map[string]interface{}{
		"edtf:inception": edtf.UNKNOWN,
		"edtf:cessation": edtf.OPEN,
	}

	for k, v := range ensure {

		_, exists := new_props[k]

		if !exists {
			new_props[k] = v
		}
	}

	parent_body, err := sfom_reader.LoadBytesFromID(ctx, r, parent_id)

	if err != nil {
		return -1, fmt.Errorf("Failed to load parent feature %d, %v", parent_id, err)
	}

	parent_f, err := geojson.UnmarshalFeature(parent_body)

	if err != nil {
		return -1, fmt.Errorf("Failed to parse parent feature %d, %v", parent_id, err)
	}

	parent_props := parent_f.Properties

	sfom_pt := new_props["sfomuseum:placetype"].(string)
	var pt string

	switch sfom_pt {
	case "bart":
		pt = "venue"
	case "terminal":
		pt = "wing"
	case "boardingarea":
		pt = "concourse"
	case "commonarea":
		pt = "concourse"
	case "observationdeck":
		pt = "arcade"
	case "gate":
		pt = "venue"
	case "checkpoint":
		pt = "venue"
	default:
		return -1, fmt.Errorf("Unknown sfomuseum:placetype '%s'", sfom_pt)
	}

	new_props["wof:parent_id"] = parent_id
	new_props["wof:hierarchy"] = parent_props["wof:hierarchy"]

	new_props["mz:is_current"] = 1
	new_props["iso:country"] = "US"

	new_props["wof:name"] = new_props["sfomuseum:name"].(string)
	new_props["wof:placetype"] = pt
	new_props["wof:placetype_alt"] = []string{
		sfom_pt,
	}

	new_props["sfomuseum:is_sfo"] = 1

	new_props["wof:supersedes"] = make([]int64, 0)
	new_props["wof:superseded_by"] = make([]int64, 0)

	new_props["src:geom"] = "sfo"

	date := new_props["edtf:inception"].(string)
	new_props["wof:label"] = fmt.Sprintf("%s (%s)", new_props["wof:name"].(string), date)

	new_f.Properties = new_props

	enc_f, err := new_f.MarshalJSON()

	if err != nil {
		return -1, err
	}

	new_id, err := sfom_writer.WriteFeatureBytes(ctx, wr, enc_f)

	if err != nil {
		return -1, err
	}

	return new_id, nil
}
