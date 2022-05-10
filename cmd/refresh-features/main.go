package main

import (
	_ "github.com/whosonfirst/go-reader-github"
	_ "github.com/whosonfirst/go-reader-http"
	_ "github.com/whosonfirst/go-reader-whosonfirst-data"	
)

import (
	"context"
	"flag"
	"fmt"
	_ "github.com/sfomuseum/go-sfomuseum-export/v2"
	"github.com/sfomuseum/go-sfomuseum-whosonfirst/custom"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-fetch"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer"
	"io"
	"log"
	"net/url"
)

func main() {

	iterator_uri := flag.String("data-iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/v2 URI")
	iterator_source := flag.String("data-iterator-source", "/usr/local/data/sfomuseum-data-whosonfirst", "...")

	wof_reader_uri := flag.String("whosonfirst-reader-uri", "whosonfirst-data://", "A valid whosonfirst/go-reader URI.")

	data_reader_uri := flag.String("data-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/data", "A valid whosonfirst/go-reader URI.")
	properties_reader_uri := flag.String("properties-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-reader URI.")

	data_writer_uri := flag.String("data-writer-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/data", "A valid whosonfirst/go-writer URI.")
	properties_writer_uri := flag.String("properties-writer-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-writer URI.")

	data_exporter_uri := flag.String("data-exporter-uri", "sfomuseum://", "A valid whosonfirst/go-whosonfirst-export URI.")

	retries := flag.Int("retries", 3, "The maximum number of attempts to try fetching a record.")
	max_clients := flag.Int("max-clients", 10, "The maximum number of concurrent requests for multiple Who's On First records.")

	user_agent := flag.String("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:10.0) Gecko/20100101 Firefox/10.0", "An optional user-agent to append to the -whosonfirst-reader-uri flag")

	flag.Parse()

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
		log.Fatalf("Failed to create new fetcher, %v", err)
	}

	sfom_opts := &custom.SFOMuseumPropertiesOptions{
		DataReader:       data_r,
		DataWriter:       data_wr,
		DataExporter:     data_ex,
		PropertiesReader: props_r,
		PropertiesWriter: props_wr,
	}

	belongs_to := []string{
		"region",
		"country",
	}

	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		id, _, err := uri.ParseURI(path)

		if err != nil {
			return fmt.Errorf("Failed to derive ID from %s, %w", path, err)
		}

		// START OF put me in a function

		to_fetch := []int64{id}

		_, err = fetcher.FetchIDs(ctx, to_fetch, belongs_to...)

		if err != nil {

			fmt.Printf("Failed to fetch %d (%s), %v", id, path, err)
			return nil
			
			// return fmt.Errorf("Failed to fetch %d (%s), %w", id, path, err)
		}

		err = custom.ApplySFOMuseumProperties(ctx, sfom_opts, id)

		if err != nil {
			return fmt.Errorf("Failed to apply SFO Museum properties for %d (%s), %v", id, path, err)
		}

		// END OF put me in a function

		return nil
	}

	iter, err := iterator.NewIterator(ctx, *iterator_uri, iter_cb)

	if err != nil {
		log.Fatalf("Failed to create new iterator, %v", err)
	}

	err = iter.IterateURIs(ctx, *iterator_source)

	if err != nil {
		log.Fatalf("Failed to iterate URIs, %v", err)
	}

}
