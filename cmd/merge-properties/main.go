// merge-properties iterates over a collection of Who's On First records and merges custom properties.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-sfomuseum-whosonfirst/custom"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer"
	"io"
	"log"
	"os"
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

	iter_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		id, uri_args, err := uri.ParseURI(path)

		if err != nil {
			return fmt.Errorf("Failed to parse '%s', %v", path, err)
		}

		if uri_args.IsAlternate {
			return nil
		}

		err = custom.MergeCustomProperties(ctx, props_r, r, wr, id)

		if err != nil {
			return fmt.Errorf("Failed to merge properties for %d, %w", id, err)
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
