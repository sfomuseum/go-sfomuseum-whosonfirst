// merge-properties iterates over a collection of Who's On First records and merges custom properties.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/sfomuseum/go-sfomuseum-export/v2"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer"
	"io"
	"log"
	"path/filepath"
)

func main() {

	iter_uri := flag.String("iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/v2 URI.")

	properties_reader_uri := flag.String("properties-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-reader.Reader URI.")
	writer_uri := flag.String("writer-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/data", "A valid whosonfirst/go-writer.Writer URI.")

	exporter_uri := flag.String("exporter-uri", "sfomuseum://", "A valid whosonfirst/go-export/v2.Exporter URI.")

	flag.Parse()

	uris := flag.Args()

	ctx := context.Background()

	wr, err := writer.NewWriter(ctx, *writer_uri)

	if err != nil {
		log.Fatalf("Failed to create new writer, %v", err)
	}

	props_r, err := reader.NewReader(ctx, *properties_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create reader, %v", err)
	}

	ex, err := export.NewExporter(ctx, *exporter_uri)

	if err != nil {
		log.Fatalf("Failed to create exporter, %v", err)
	}

	iter_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		id, uri_args, err := uri.ParseURI(path)

		if err != nil {
			return fmt.Errorf("Failed to parse '%s', %v", path, err)
		}

		if uri_args.IsAlternate {
			return nil
		}

		props_tree, err := uri.Id2Path(id)

		if err != nil {
			return fmt.Errorf("Failed to derive path for '%d', %v", id, err)
		}

		props_fname := fmt.Sprintf("%d.json", id)
		props_path := filepath.Join(props_tree, props_fname)

		props_fh, err := props_r.Read(ctx, props_path)

		if err != nil {
			return fmt.Errorf("Failed to read '%s', %v", props_path, err)
		}

		var props_map map[string]interface{}

		dec := json.NewDecoder(props_fh)
		err = dec.Decode(&props_map)

		if err != nil {
			return fmt.Errorf("Failed to decode properties map, %v", err)
		}

		fq_props := make(map[string]interface{})

		for k, v := range props_map {
			fq_k := fmt.Sprintf("properties.%s", k)
			fq_props[fq_k] = v
		}

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read feature body, %v", err)
		}

		changed, new_body, err := export.AssignPropertiesIfChanged(ctx, body, fq_props)

		if err != nil {
			return fmt.Errorf("Failed to assign properties for '%s', %v", path, err)
		}

		if !changed {
			return nil
		}

		new_body, err = ex.Export(ctx, new_body)

		if err != nil {
			return fmt.Errorf("Failed to export feature ID '%d', %v", id, err)
		}

		br := bytes.NewReader(new_body)
		cl, err := ioutil.NewReadSeekCloser(br)

		if err != nil {
			return fmt.Errorf("Failed to create new ReadSeekCloser for '%d', %v", id, err)
		}

		rel_path, err := uri.Id2RelPath(id)

		if err != nil {
			return fmt.Errorf("Failed to derive relative path for '%d', %v", id, err)
		}
		
		_, err = wr.Write(ctx, rel_path, cl)

		if err != nil {
			return fmt.Errorf("Failed to write '%s', %v", rel_path, err)
		}

		log.Printf("Merge %s\n", rel_path)
		return nil
	}

	iter, err := iterator.NewIterator(ctx, *iter_uri, iter_cb)

	if err != nil {
		log.Fatalf("Failed to create new iterator, %v", err)
	}

	err = iter.IterateURIs(ctx, uris...)

	if err != nil {
		log.Fatalf("Failed to iterate URIs, %v", err)
	}

}
