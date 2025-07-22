// ensure-properties iterates over a collection of Who's On First records and ensures that there is a corresponding properties JSON file.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/sfomuseum/go-sfomuseum-whosonfirst/custom"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader/v2"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v3"
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

	iter, err := iterate.NewIterator(ctx, *iter_uri)

	if err != nil {
		log.Fatalf("Failed to create new iterator, %v", err)
	}

	for rec, err := range iter.Iterate(ctx, uris...) {

		if err != nil {
			log.Fatal(err)
		}

		defer rec.Body.Close()

		id, uri_args, err := uri.ParseURI(rec.Path)

		if err != nil {
			return fmt.Errorf("Failed to parse '%s', %v", rec.Path, err)
		}

		if uri_args.IsAlternate {
			continue
		}

		body, err := io.ReadAll(fh)

		if err != nil {
			log.Fatalf("Failed to read '%s', %v", rec.Path, err)
		}

		props_map, err := custom.EnsureCustomProperties(ctx, props_r, props_wr, id)

		if err != nil {
			log.Fatalf("Failed to load custom properties for for '%d', %v", id, err)
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
				log.Fatalf("Failed to derive wof:placetype for '%s'", rec.Path)
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
			log.Fatalf("Failed to write custom properties for %d, %v", id, err)
		}

		return nil
	}

}
