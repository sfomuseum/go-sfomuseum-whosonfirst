// merge-properties iterates over a collection of Who's On First records and merges custom properties.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-iterate/emitter"
	"github.com/whosonfirst/go-whosonfirst-iterate/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer"
	"io"
	"log"
	"path/filepath"
)

func main() {

	iter_uri := flag.String("iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/emitter URI.")

	properties_reader_uri := flag.String("properties-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-reader.Reader URI.")
	properties_writer_uri := flag.String("properties-writer-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-writer.Writer URI.")

	flag.Parse()

	uris := flag.Args()

	ctx := context.Background()

	props_wr, err := writer.NewWriter(ctx, *properties_writer_uri)

	if err != nil {
		log.Fatalf("Failed to create new writer, %v", err)
	}

	props_r, err := reader.NewReader(ctx, *properties_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create reader, %v", err)
	}

	iter_cb := func(ctx context.Context, fh io.ReadSeeker, args ...interface{}) error {

		path, err := emitter.PathForContext(ctx)

		if err != nil {
			return fmt.Errorf("Failed to derive path for context, %v", err)
		}

		id, uri_args, err := uri.ParseURI(path)

		if err != nil {
			return fmt.Errorf("Failed to parse '%s', %v", path, err)
		}

		if uri_args.IsAlternate {
			return nil
		}

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read '%s', %v", path, err)
		}

		props_tree, err := uri.Id2Path(id)

		if err != nil {
			return fmt.Errorf("Failed to derive path for '%d', %v", id, err)
		}

		var props_map map[string]interface{}
		has_updates := false

		props_fname := fmt.Sprintf("%d.json", id)
		props_path := filepath.Join(props_tree, props_fname)

		props_fh, err := props_r.Read(ctx, props_path)

		if err == nil {

			dec := json.NewDecoder(props_fh)
			err = dec.Decode(&props_map)

			if err != nil {
				return fmt.Errorf("Failed to decode properties map, %v", err)
			}

		} else {

			props_map = make(map[string]interface{})
			has_updates = true
		}

		_, repo_ok := props_map["wof:repo"]

		if !repo_ok {
			props_map["wof:repo"] = "sfomuseum-data-whosonfirst"
			has_updates = true
		}

		_, placetype_ok := props_map["sfomuseum:placetype"]

		if !placetype_ok {
			pt_rsp := gjson.GetBytes(body, "properties.wof:placetype")

			if !pt_rsp.Exists() {
				return fmt.Errorf("Failed to derive wof:placetype for '%s'", path)
			}

			props_map["sfomuseum:placetype"] = pt_rsp.String()
			has_updates = true
		}

		if !has_updates {
			return nil
		}

		props_body, err := json.Marshal(props_map)

		if err != nil {
			return fmt.Errorf("Failed to marshal '%s', %v", props_path, err)
		}

		props_body = pretty.Pretty(props_body)

		br := bytes.NewReader(props_body)
		cl, err := ioutil.NewReadSeekCloser(br)

		if err != nil {
			return fmt.Errorf("Failed to create new ReadSeekCloser for '%s', %v", props_path, err)
		}

		_, err = props_wr.Write(ctx, props_path, cl)

		if err != nil {
			return fmt.Errorf("Failed to write '%s', %v", props_path, err)
		}

		log.Printf("Write %s\n", props_path)
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
