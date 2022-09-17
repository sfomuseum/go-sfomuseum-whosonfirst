// package custom provides methods for working with custom properties to be applied to Who's On First records.
// This (and the corresponding tools in cmd) should be migrated in to a standalone whosonfirst/go-whosonfirst-custom package.
package custom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer/v3"
	"github.com/tidwall/pretty"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v3"
	"io"
	"path/filepath"
)

// Id2RelPath will return a relative path (URI) for id.
func Id2RelPath(id int64) (string, error) {

	props_tree, err := uri.Id2Path(id)

	if err != nil {
		return "", fmt.Errorf("Failed to derive path for '%d', %v", id, err)
	}

	props_fname := fmt.Sprintf("%d.json", id)
	rel_path := filepath.Join(props_tree, props_fname)

	return rel_path, nil
}

// ReadCustomProperties will return a properties map for id, reading the raw data from r.
func ReadCustomProperties(ctx context.Context, r reader.Reader, id int64) (map[string]interface{}, error) {

	props_path, err := Id2RelPath(id)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive path for '%d', %w", id, err)
	}

	props_fh, err := r.Read(ctx, props_path)

	if err != nil {
		return nil, fmt.Errorf("Failed to read custom properties for %s, %w", props_path, err)
	}

	var props_map map[string]interface{}

	dec := json.NewDecoder(props_fh)
	err = dec.Decode(&props_map)

	if err != nil {
		return nil, fmt.Errorf("Failed to decode custom properties for %d, %w", id, err)
	}

	return props_map, nil
}

// WriteCustomProperties writes props_map to a relative path derived from id, using wr.
func WriteCustomProperties(ctx context.Context, wr writer.Writer, id int64, props_map map[string]interface{}) error {

	props_path, err := Id2RelPath(id)

	if err != nil {
		return fmt.Errorf("Failed to derive path for '%d', %w", id, err)
	}

	props_body, err := json.Marshal(props_map)

	if err != nil {
		return fmt.Errorf("Failed to marshal %d, %w", id, err)
	}

	props_body = pretty.Pretty(props_body)

	br := bytes.NewReader(props_body)
	cl, err := ioutil.NewReadSeekCloser(br)

	if err != nil {
		return fmt.Errorf("Failed to create new ReadSeekCloser for %d, %w", id, err)
	}

	_, err = wr.Write(ctx, props_path, cl)

	if err != nil {
		return fmt.Errorf("Failed to write custom properties for %d, %w", id, err)
	}

	// See this? It's important if we're reading/writing to a GitHub repo; specifically we
	// need to make sure we push (flush) the record above in order to be able to read it
	// when `ReadCustomProperties` is called.

	err = wr.Flush(ctx)

	if err != nil {
		return fmt.Errorf("Failed to flush custom properties for %d, %w", id, err)
	}

	return nil
}

// EnsureCustomProperties will ensure that a valid properties map exists for id in r. If it does not an empty properties map file will be written to wr.
func EnsureCustomProperties(ctx context.Context, r reader.Reader, wr writer.Writer, id int64) (map[string]interface{}, error) {

	props_map, err := ReadCustomProperties(ctx, r, id)

	if err == nil {
		return props_map, nil
	}

	return CreateCustomProperties(ctx, wr, id)
}

// CreateCustomProperties will create a new (empty) properties map for id, using wr.
func CreateCustomProperties(ctx context.Context, wr writer.Writer, id int64) (map[string]interface{}, error) {

	props_map := make(map[string]interface{})

	err := WriteCustomProperties(ctx, wr, id, props_map)

	if err != nil {
		return nil, fmt.Errorf("Failed to write new custom properties for %d, %w", id, err)
	}

	return props_map, nil
}

// MergeCustomProperties will merge the custom properties for id read from props_r in to a WOF record (for id) read from data_r. The merged document will
// be exported and published using data_wr.
func MergeCustomProperties(ctx context.Context, props_r reader.Reader, data_r reader.Reader, data_wr writer.Writer, id int64) error {

	data_path, err := uri.Id2RelPath(id)

	if err != nil {
		return fmt.Errorf("Failed to derive path for %d, %w", id, err)
	}

	data_fh, err := data_r.Read(ctx, data_path)

	if err != nil {
		return fmt.Errorf("Failed read data for %s, %w", data_path, err)
	}

	body, err := io.ReadAll(data_fh)

	if err != nil {
		return fmt.Errorf("Failed to read feature body, %v", err)
	}

	props_map, err := ReadCustomProperties(ctx, props_r, id)

	if err != nil {
		return fmt.Errorf("Failed to read custom properties for %d, %v", id, err)
	}

	props_map = ApplyEDTFFixes(ctx, body, props_map)

	fq_props := make(map[string]interface{})

	for k, v := range props_map {
		fq_k := fmt.Sprintf("properties.%s", k)
		fq_props[fq_k] = v
	}

	changed, new_body, err := export.AssignPropertiesIfChanged(ctx, body, fq_props)

	if err != nil {
		return fmt.Errorf("Failed to assign properties for '%s', %v", data_path, err)
	}

	if changed {

		_, err = sfom_writer.WriteBytes(ctx, data_wr, new_body)

		if err != nil {
			return fmt.Errorf("Failed to write feature for '%d', %w", id, err)
		}
	}

	return nil
}
