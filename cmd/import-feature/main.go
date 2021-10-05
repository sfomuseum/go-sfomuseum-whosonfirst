// import-feature fetches a WOF record for one or more IDs and writes them to a SFO Museum data repository (sfomuseum-data-whosonfirst)
//
// There is a certain amount of overlap with this code and the code in cmd/ensure-properties and cmd/merge-properties that should be
// reconciled. Also it is not possible to pass in custom sfomuseum: properties to append to the (SFO Museum) properties JSON files. There
// should be.
package main

import (
	_ "github.com/whosonfirst/go-reader-http"
	_ "github.com/whosonfirst/go-reader-whosonfirst-data"
)

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mitchellh/go-wordwrap"
	sfom_reader "github.com/sfomuseum/go-sfomuseum-reader"
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-fetch"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"	
	"github.com/whosonfirst/go-writer"
	"github.com/sfomuseum/go-edtf"	
	"github.com/tidwall/gjson"	
	"log"
	"os"
	"path/filepath"
	"strings"
)

type SFOMuseumProperties map[string]interface{}

func main() {

	wof_reader_uri := flag.String("whosonfirst-reader-uri", "whosonfirst-data://", "A valid whosonfirst/go-reader URI.")

	data_reader_uri := flag.String("data-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/data", "A valid whosonfirst/go-reader URI.")
	properties_reader_uri := flag.String("properties-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-reader URI.")

	data_writer_uri := flag.String("data-writer-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/data", "A valid whosonfirst/go-writer URI.")
	properties_writer_uri := flag.String("properties-writer-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-writer URI.")

	retries := flag.Int("retries", 3, "The maximum number of attempts to try fetching a record.")
	max_clients := flag.Int("max-clients", 10, "The maximum number of concurrent requests for multiple Who's On First records.")

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
		log.Fatalf("Failed to create new WOF reader, %v", err)
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
		
		fname := fmt.Sprintf("%d.json", id)
		tree, err := uri.Id2Path(id)

		if err != nil {
			log.Fatalf("Failed to derive tree for ID %d, %v", id, err)
		}

		rel_path := filepath.Join(tree, fname)

		var props map[string]interface{}

		props_fh, err := props_r.Read(ctx, rel_path)

		if err == nil {

			dec := json.NewDecoder(props_fh)
			err = dec.Decode(&props)

			if err != nil {
				log.Fatalf("Failed to decode properties for %d (%s), %v", id, rel_path, err)
			}
		} else {
			props = make(map[string]interface{})
		}

		props["wof:repo"] = "sfomuseum-data-whosonfirst"

		switch data_pt {
		case "campus":
			props["sfomuseum:placetype"] = "airport"
		case "locality":
			props["sfomuseum:placetype"] = "city"
		default:
			// pass
		}

		// To do: Other SFO Museum specific properties here
		
		var props_buf bytes.Buffer
		props_buf_wr := bufio.NewWriter(&props_buf)

		enc := json.NewEncoder(props_buf_wr)
		err = enc.Encode(props)

		if err != nil {
			log.Fatalf("Failed to encode properties for %d (%s), %v", id, rel_path, err)
		}

		props_buf_wr.Flush()

		br := bytes.NewReader(props_buf.Bytes())
		props_rsc, err := ioutil.NewReadSeekCloser(br)

		if err != nil {
			log.Fatalf("Failed to create ReadSeekCloser for %d (%s), %v", id, rel_path, err)
		}

		_, err = props_wr.Write(ctx, rel_path, props_rsc)

		// Now update the WOF record

		to_update := make(map[string]interface{})

		for k, v := range props {
			path := fmt.Sprintf("properties.%s", k)
			to_update[path] = v
		}

		// START OF account for known-bunk EDTF values in WOF records
		// This should be made in to a package method or something

		data_props := gjson.GetBytes(data_body, "properties")
		
		for k, v := range data_props.Map() {

			if !strings.HasPrefix(k, "edtf:") {
	                        continue
			}

	                path := fmt.Sprintf("properties.%s", k)

			switch v.String() {
		        case "open":

				to_update[path] = edtf.OPEN

		        case "uuuu":

				to_update[path] = edtf.UNSPECIFIED				

			default:
                                // pass
			}
		}

		// END OF account for known-bunk EDTF values in WOF records		
		
		has_changed, data_body, err := export.AssignPropertiesIfChanged(ctx, data_body, to_update)

		if err != nil {
			log.Fatalf("Failed to assign SFOM properties to WOF record %d, %v", id, err)
		}

		if has_changed {
			_, err = sfom_writer.WriteFeatureBytes(ctx, data_wr, data_body)

			if err != nil {
				log.Fatalf("Failed to write WOF record %d, %v", id, err)
			}
		}

		// END OF put me in a package method or something		
	}
}
