package main

import (
	_ "github.com/whosonfirst/go-reader-github"
	_ "github.com/whosonfirst/go-reader-http"
)

import (
	"context"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-sfomuseum-whosonfirst/custom"
	wof_import "github.com/sfomuseum/go-sfomuseum-whosonfirst/import"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-fetch"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v2"
	"io"
	"log"
	"net/url"
	"sync"
)

func main() {

	iterator_uri := flag.String("iterator-uri", "repo://", "")

	wof_reader_uri := flag.String("whosonfirst-reader-uri", "https://data.whosonfirst.org/", "A valid whosonfirst/go-reader URI.")

	data_reader_uri := flag.String("data-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/data", "A valid whosonfirst/go-reader URI.")
	properties_reader_uri := flag.String("properties-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-reader URI.")

	data_writer_uri := flag.String("data-writer-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/data", "A valid whosonfirst/go-writer URI.")
	properties_writer_uri := flag.String("properties-writer-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-writer URI.")

	retries := flag.Int("retries", 3, "The maximum number of attempts to try fetching a record.")
	max_clients := flag.Int("max-clients", 10, "The maximum number of concurrent requests for multiple Who's On First records.")

	user_agent := flag.String("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:10.0) Gecko/20100101 Firefox/10.0", "An optional user-agent to append to the -whosonfirst-reader-uri flag")

	flag.Parse()

	iterator_sources := flag.Args()

	ctx := context.Background()

	if *user_agent != "" {

		wof_u, err := url.Parse(*wof_reader_uri)

		if err != nil {
			log.Fatalf("Failed to parse (WOF) reader URI, %v", err)
		}

		q := wof_u.Query()
		q.Set("user-agent", *user_agent)

		wof_u.RawQuery = q.Encode()
		*wof_reader_uri = wof_u.String()
	}

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
		log.Fatalf("Failed to create new data writer, %v", err)
	}

	props_wr, err := writer.NewWriter(ctx, *properties_writer_uri)

	if err != nil {
		log.Fatalf("Failed to create new properties writer, %v", err)
	}

	query_paths := []string{
		"properties.sfomuseum:flightcover_address_from",
		"properties.sfomuseum:flightcover_address_to",
		"properties.sfomuseum:flightcover_postmark_sent",
		"properties.sfomuseum:flightcover_postmark_received",
	}

	features_map := new(sync.Map)

	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		_, uri_args, err := uri.ParseURI(path)

		if err != nil {
			return fmt.Errorf("Failed to parse %s, %w", path, err)
		}

		if uri_args.IsAlternate {
			return nil
		}

		body, err := io.ReadAll(r)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %w", path, err)
		}

		done_ch := make(chan bool)

		for _, p := range query_paths {

			go func(ctx context.Context, p string) {

				rsp := gjson.GetBytes(body, p)

				if rsp.Exists() {

					for _, r := range rsp.Array() {
						features_map.Store(r.Int(), true)
					}
				}

				done_ch <- true

			}(ctx, p)
		}

		remaining := len(query_paths)

		for remaining > 0 {
			select {
			case <-done_ch:
				remaining -= 1
			}
		}

		return nil
	}

	iter, err := iterator.NewIterator(ctx, *iterator_uri, iter_cb)

	if err != nil {
		log.Fatalf("Failed to create new iterator, %w", err)
	}

	err = iter.IterateURIs(ctx, iterator_sources...)

	if err != nil {
		log.Fatalf("Failed to iterate URIs, %w", err)
	}

	//

	feature_ids := make([]int64, 0)

	features_map.Range(func(k interface{}, v interface{}) bool {
		id := k.(int64)
		feature_ids = append(feature_ids, id)
		return true
	})

	//

	fetcher_opts, err := fetch.DefaultOptions()

	if err != nil {
		log.Fatalf("Failed to create fetch options, %v", err)
	}

	fetcher_opts.Retries = *retries
	fetcher_opts.MaxClients = *max_clients

	fetcher, err := fetch.NewFetcher(ctx, wof_r, data_wr, fetcher_opts)

	if err != nil {
		log.Fatalf("Failed to create new fetcher, %v", err)
	}

	sfom_opts := &custom.SFOMuseumPropertiesOptions{
		DataReader:       data_r,
		DataWriter:       data_wr,
		PropertiesReader: props_r,
		PropertiesWriter: props_wr,
	}

	belongs_to := []string{
		"region",
		"country",
	}

	import_opts := &wof_import.ImportFeatureOptions{
		Fetcher:           fetcher,
		PropertiesOptions: sfom_opts,
		BelongsTo:         belongs_to,
	}

	err = wof_import.ImportFeatures(ctx, import_opts, feature_ids...)

	if err != nil {
		log.Fatalf("Failed to import features, %v", err)
	}

}
