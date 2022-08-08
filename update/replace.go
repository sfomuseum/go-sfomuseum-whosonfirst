package update

import (
	"context"
	"fmt"
	"github.com/paulmach/orb/geojson"
	"github.com/sfomuseum/go-edtf"
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer/v2"
	"github.com/whosonfirst/go-reader"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	"github.com/whosonfirst/go-writer/v2"
	"strings"
)

// Replace (or superseded) a new Who's On First record and write it to 'wr'. The signature for this method is a bit of a mess right now
// and reflects the specific concerns of the go-sfomuseum-gis/cmd/sfo-export-complex tool where	it was first written. This method shares many
// of the same concerns as CreateWithFeature and the two methods should, where possible, be reconciled.
func ReplaceWithSelf(ctx context.Context, r reader.Reader, wr writer.Writer, old_id int64, new_parent_id int64, date string) (int64, error) {

	tmp_f, err := wof_reader.LoadFeature(ctx, r, old_id)

	if err != nil {
		return -1, err
	}

	return ReplaceWithFeature(ctx, r, wr, tmp_f, old_id, new_parent_id, date)
}

// Replace (or superseded) a new Who's On First record from a geojson.Feature instance and write it to 'wr'. See notes about method signatures
// in the ReplaceWithSelf method.
func ReplaceWithFeature(ctx context.Context, r reader.Reader, wr writer.Writer, tmp_f *geojson.Feature, old_id int64, new_parent_id int64, date string) (int64, error) {

	old_f, err := wof_reader.LoadFeature(ctx, r, old_id)

	if err != nil {
		return -1, fmt.Errorf("Failed to load %d, %v", old_id, err)
	}

	new_f := geojson.NewFeature(tmp_f.Geometry)

	old_props := old_f.Properties // the last known good record for this place
	new_props := new_f.Properties // the new record being produce
	tmp_props := tmp_f.Properties // the record exported from ArcGIS

	for key, value := range old_props {

		switch key {
		case "wof:id", "wof:created", "wof:lastmodified":
			// pass
		default:
			new_props[key] = value
		}
	}

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

	// update new properties

	if new_parent_id == -1 {
		// pass
	} else {

		parent_f, err := wof_reader.LoadFeature(ctx, r, new_parent_id)

		if err != nil {
			return -1, fmt.Errorf("Failed to load parent %d, %v", new_parent_id, err)
		}

		parent_props := parent_f.Properties

		new_props["wof:parent_id"] = new_parent_id
		new_props["wof:hierarchy"] = parent_props["wof:hierarchy"]
	}

	// TO DO: parent id and hierarchy...

	supersedes := []int64{
		old_id,
	}

	new_props["wof:supersedes"] = supersedes

	new_props["edtf:inception"] = date
	new_props["edtf:cessation"] = edtf.OPEN
	new_props["src:geom"] = "sfo"

	new_props["wof:label"] = fmt.Sprintf("%s (%s)", new_props["wof:name"].(string), date)

	new_props["mz:is_current"] = 1

	//

	new_f.Properties = new_props

	new_id, err := sfom_writer.WriteFeature(ctx, wr, new_f)

	if err != nil {
		return -1, err
	}

	// update old record here

	superseded_by := make([]int64, 0)

	old_superseded_by, ok := old_props["wof:superseded_by"]

	if ok {

		for _, i := range old_superseded_by.([]interface{}) {
			superseded_by = append(superseded_by, int64(i.(float64)))
		}

	}

	superseded_by = append(superseded_by, new_id)

	old_props["edtf:cessation"] = date
	old_props["mz:is_current"] = 0
	old_props["wof:superseded_by"] = superseded_by

	old_f.Properties = old_props

	_, err = sfom_writer.WriteFeature(ctx, wr, old_f)

	if err != nil {
		return -1, err
	}

	return new_id, err
}
