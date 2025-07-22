// merge-properties iterates over a collection of Who's On First records and merges custom properties.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sfomuseum/go-sfomuseum-whosonfirst/custom"
	"github.com/whosonfirst/go-reader/v2"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v3"
)

func main() {

	iter_uri := flag.String("iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/v2 URI.")

	properties_reader_uri := flag.String("properties-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-reader.Reader URI.")

	reader_uri := flag.String("reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/data", "A valid whosonfirst/go-reader.Reader URI.")
	writer_uri := flag.String("writer-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/data", "A valid whosonfirst/go-writer.Writer URI.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "merge-properties iterates over a collection of Who's On First records and merges custom properties.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options] record(N) record(N)\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	uris := flag.Args()

	ctx := context.Background()

	wr, err := writer.NewWriter(ctx, *writer_uri)

	if err != nil {
		log.Fatalf("Failed to create new writer, %v", err)
	}

	r, err := reader.NewReader(ctx, *reader_uri)

	if err != nil {
		log.Fatalf("Failed to create (data) reader, %v", err)
	}

	props_r, err := reader.NewReader(ctx, *properties_reader_uri)

	if err != nil {
		log.Fatalf("Failed to create (properties) reader, %v", err)
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
			log.Fatalf("Failed to parse '%s', %v", rec.Path, err)
		}

		if uri_args.IsAlternate {
			continue
		}

		err = custom.MergeCustomProperties(ctx, props_r, r, wr, id)

		if err != nil {
			log.Fatalf("Failed to merge properties for %d, %v", id, err)
		}
	}

}
