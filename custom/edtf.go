package custom

import (
	"context"
	"github.com/sfomuseum/go-edtf"
	"github.com/tidwall/gjson"
	"strings"
)

// ApplyEDTFFixes applies EDTF 2019 updates to 'props_maps' (derived from 'body') if necessary.
func ApplyEDTFFixes(ctx context.Context, body []byte, props_map map[string]interface{}) map[string]interface{} {

	data_props := gjson.GetBytes(body, "properties")

	for k, v := range data_props.Map() {

		if !strings.HasPrefix(k, "edtf:") {
			continue
		}

		path := k

		switch v.String() {
		case "open":
			props_map[path] = edtf.OPEN
		case "uuuu":
			props_map[path] = edtf.UNSPECIFIED
		default:
			// pass
		}
	}

	return props_map
}
