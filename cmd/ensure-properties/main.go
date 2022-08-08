// ensure-properties iterates over a collection of Who's On First records and ensures that there is a corresponding properties JSON file.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-sfomuseum-whosonfirst/custom"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v2"
	"io"
	"log"
	"os"
)

func main() {

	iter_uri := flag.String("iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/v2 URI.")

	properties_reader_uri := flag.String("properties-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-reader.Reader URI.")
	properties_writer_uri := flag.String("properties-writer-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-writer.Writer URI.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ensure-properties iterates over a collection of Who's On First records and ensures that there is a corresponding properties JSON file.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options] uri(N) uri(N)\n", os.Args[0])
		flag.PrintDefaults()
	}

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

	iter_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

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

		props_map, err := custom.EnsureCustomProperties(ctx, props_r, props_wr, id)

		if err != nil {
			return fmt.Errorf("Failed to load custom properties for for '%d', %v", id, err)
		}

		has_updates := false

		_, repo_ok := props_map["sfomuseum:repo"]

		if !repo_ok {
			props_map["sfomuseum:repo"] = "sfomuseum-data-whosonfirst"
			has_updates = true
		}

		_, placetype_ok := props_map["sfomuseum:placetype"]

		if !placetype_ok {

			pt_rsp := gjson.GetBytes(body, "properties.wof:placetype")

			if !pt_rsp.Exists() {
				return fmt.Errorf("Failed to derive wof:placetype for '%s'", path)
			}

			switch pt_rsp.String() {
			case "campus":
				props_map["sfomuseum:placetype"] = "airport"
			case "locality":
				props_map["sfomuseum:placetype"] = "city"
			default:
				props_map["sfomuseum:placetype"] = pt_rsp.String()
			}

			has_updates = true
		}

		if !has_updates {
			return nil
		}

		err = custom.WriteCustomProperties(ctx, props_wr, id, props_map)

		if err != nil {
			return fmt.Errorf("Failed to write custom properties for %d, %v", id, err)
		}

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
