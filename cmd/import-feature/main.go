// import-feature fetches a WOF record for one or more IDs and writes them to a SFO Museum data repository (sfomuseum-data-whosonfirst)
//
// There is a certain amount of overlap with this code and the code in cmd/ensure-properties and cmd/merge-properties that should be
// reconciled. Also it is not possible to pass in custom sfomuseum: properties to append to the (SFO Museum) properties JSON files. There
// should be.
package main

import (
	_ "github.com/whosonfirst/go-reader-github"
	_ "github.com/whosonfirst/go-reader-http"
)

import (
	"context"
	"flag"
	"fmt"
	"github.com/mitchellh/go-wordwrap"
	"github.com/sfomuseum/go-flags/multi"
	_ "github.com/sfomuseum/go-sfomuseum-export/v2"
	sfom_reader "github.com/sfomuseum/go-sfomuseum-reader"
	"github.com/sfomuseum/go-sfomuseum-whosonfirst/custom"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-fetch"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer"
	"log"
	"os"
	"strings"
)

func main() {

	wof_reader_uri := flag.String("whosonfirst-reader-uri", "https://data.whosonfirst.org/", "A valid whosonfirst/go-reader URI.")

	data_reader_uri := flag.String("data-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/data", "A valid whosonfirst/go-reader URI.")
	properties_reader_uri := flag.String("properties-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-reader URI.")

	data_writer_uri := flag.String("data-writer-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/data", "A valid whosonfirst/go-writer URI.")
	properties_writer_uri := flag.String("properties-writer-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-writer URI.")

	data_exporter_uri := flag.String("data-exporter-uri", "sfomuseum://", "A valid whosonfirst/go-whosonfirst-export URI.")

	retries := flag.Int("retries", 3, "The maximum number of attempts to try fetching a record.")
	max_clients := flag.Int("max-clients", 10, "The maximum number of concurrent requests for multiple Who's On First records.")

	var str_properties multi.KeyValueString
	flag.Var(&str_properties, "string-property", "One or more {KEY}={VALUE} flags where {KEY} is a valid tidwall/gjson path and {VALUE} is a string value.")

	var int_properties multi.KeyValueInt64
	flag.Var(&int_properties, "int-property", "One or more {KEY}={VALUE} flags where {KEY} is a valid tidwall/gjson path and {VALUE} is a int(64) value.")

	// var float_properties multi.KeyValueFloat64
	// flag.Var(&float_properties, "float-property", "One or more {KEY}={VALUE} flags where {KEY} is a valid tidwall/gjson path and {VALUE} is a float(64) value.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Fetch one or more Who's on First records and, optionally, their ancestors.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [options] [path1 path2 ... pathN]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nNotes:\n\n")
		fmt.Fprintf(os.Stderr, wordwrap.WrapString("pathN may be any valid Who's On First ID or URI that can be parsed by the go-whosonfirst-uri package.\n\n", 80))
	}

	flag.Parse()

	ctx := context.Background()

	wof_r, err := reader.NewReader(ctx, *wof_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create new WOF reader for '%s', %v", *wof_reader_uri, err)
	}

	data_r, err := reader.NewReader(ctx, *data_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create new data reader, %v", err)
	}

	props_r, err := reader.NewReader(ctx, *properties_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create new properties reader, %v", err)
	}

	data_wr, err := writer.NewWriter(ctx, *data_writer_uri)

	if err != nil {
		log.Fatal("Failed to create new data writer, %v", err)
	}

	props_wr, err := writer.NewWriter(ctx, *properties_writer_uri)

	if err != nil {
		log.Fatal("Failed to create new properties writer, %v", err)
	}

	data_ex, err := export.NewExporter(ctx, *data_exporter_uri)

	if err != nil {
		log.Fatalf("Failed to create new exporter, %v", err)
	}

	fetcher_opts, err := fetch.DefaultOptions()

	if err != nil {
		log.Fatal("Failed to create fetch options, %v", err)
	}

	fetcher_opts.Retries = *retries
	fetcher_opts.MaxClients = *max_clients

	fetcher, err := fetch.NewFetcher(ctx, wof_r, data_wr, fetcher_opts)

	if err != nil {
		log.Fatal("Failed to create new fetcher, %v", err)
	}

	uris := flag.Args()
	ids := make([]int64, 0)

	for _, raw := range uris {

		id, _, err := uri.ParseURI(raw)

		if err != nil {
			log.Fatalf("Unable to parse URI '%s', %v", raw, err)
		}

		ids = append(ids, id)
	}

	belongs_to := []string{
		"region",
		"country",
	}

	fetched_ids, err := fetcher.FetchIDs(ctx, ids, belongs_to...)

	if err != nil {
		log.Fatalf("Failed to fetch IDs, %v", err)
	}

	for _, id := range fetched_ids {

		// START OF put me in a package method or something

		data_body, err := sfom_reader.LoadBytesFromID(ctx, data_r, id)

		if err != nil {
			log.Fatalf("Failed to read %d, %v", id, err)
		}

		data_pt, err := properties.Placetype(data_body)

		if err != nil {
			log.Fatalf("Failed to derive placetype for %d, %v", id, err)
		}

		props, err := custom.EnsureCustomProperties(ctx, props_r, props_wr, id)

		if err != nil {
			log.Fatalf("Failed to read custom properties for %d, %v", id, err)
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

		cli_props := false

		for _, i := range ids {

			if i == id {
				cli_props = true
				break
			}
		}

		if cli_props {

			for _, p := range str_properties {
				path := p.Key()
				value := p.Value()

				path = strings.Replace(path, "properties.", "", 1)
				props[path] = value
			}

			for _, p := range int_properties {
				path := p.Key()
				value := p.Value()

				path = strings.Replace(path, "properties.", "", 1)
				props[path] = value
			}
		}
		props = custom.ApplyEDTFFixes(ctx, data_body, props)

		err = custom.WriteCustomProperties(ctx, props_wr, id, props)

		if err != nil {
			log.Fatalf("Failed to write custom properties for %d, %v", id, err)
		}

		err = custom.MergeCustomProperties(ctx, props_r, data_r, data_wr, data_ex, id)

		if err != nil {
			log.Fatalf("Failed to merge custom properties for %d, %v", id, err)
		}

	}
}
