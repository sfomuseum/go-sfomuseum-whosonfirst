// import-feature fetches a WOF record for one or more IDs and writes them to a SFO Museum data repository (sfomuseum-data-whosonfirst)
//
// There is a certain amount of overlap with this code and the code in cmd/ensure-properties and cmd/merge-properties that should be
// reconciled. Also it is not possible to pass in custom sfomuseum: properties to append to the (SFO Museum) properties JSON files. There
// should be.
package main

import (
	_ "github.com/whosonfirst/go-reader-github"
	_ "github.com/whosonfirst/go-reader-http"
	_ "gocloud.dev/runtimevar/awsparamstore"
	_ "gocloud.dev/runtimevar/constantvar"
	_ "gocloud.dev/runtimevar/filevar"
)

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mitchellh/go-wordwrap"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
	"github.com/sfomuseum/go-sfomuseum-whosonfirst/custom"
	wof_import "github.com/sfomuseum/go-sfomuseum-whosonfirst/import"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-fetch"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v2"
	gh_writer "github.com/whosonfirst/go-writer-github/v2"
	"log"
	"net/url"
	"os"
	"strings"
)

func main() {

	fs := flagset.NewFlagSet("feature")

	wof_reader_uri := fs.String("whosonfirst-reader-uri", "https://data.whosonfirst.org/", "A valid whosonfirst/go-reader URI.")

	// data_reader_uri := fs.String("data-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/data", "A valid whosonfirst/go-reader URI.")
	// properties_reader_uri := fs.String("properties-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-reader URI.")

	// data_writer_uri := fs.String("data-writer-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/data", "A valid whosonfirst/go-writer URI.")
	// properties_writer_uri := fs.String("properties-writer-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties", "A valid whosonfirst/go-writer URI.")

	data_reader_uri := fs.String("data-reader-uri", "github://sfomuseum-data/sfomuseum-data-whosonfirst?prefix=data&branch=data", "A valid whosonfirst/go-reader URI.")
	
	properties_reader_uri := fs.String("properties-reader-uri", "github://sfomuseum-data/sfomuseum-data-whosonfirst?prefix=properties&branch=props", "A valid whosonfirst/go-reader URI.")
	
	// data_writer_uri := fs.String("data-writer-uri", "githubapi-tree://sfomuseum-data/sfomuseum-data-whosonfirst?prefix=data&access_token={access_token}&email=sfomuseumbot@localhost&description=update%20features&to-branch=test", "A valid whosonfirst/go-writer URI.")

	// properties_writer_uri := fs.String("properties-writer-uri", "githubapi-tree://sfomuseum-data/sfomuseum-data-whosonfirst?prefix=properties&access_token={access_token}&email=sfomuseumbot@localhost&description=update%20properties&to-branch=test", "A valid whosonfirst/go-writer URI.")

	data_writer_uri := fs.String("data-writer-uri", "githubapi-branch://sfomuseum-data/sfomuseum-data-whosonfirst?prefix=data&access_token={access_token}&email=sfomuseumbot@localhost&description=update%20features&to-branch=data&merge=true&remove-on-merge=true", "A valid whosonfirst/go-writer URI.")

	properties_writer_uri := fs.String("properties-writer-uri", "githubapi-branch://sfomuseum-data/sfomuseum-data-whosonfirst?prefix=properties&access_token={access_token}&email=sfomuseumbot@localhost&description=update%20properties&to-branch=props&merge=true&remove-on-merge=true", "A valid whosonfirst/go-writer URI.")
	
	token_uri := fs.String("access-token-uri", "", "")

	retries := fs.Int("retries", 3, "The maximum number of attempts to try fetching a record.")
	max_clients := fs.Int("max-clients", 10, "The maximum number of concurrent requests for multiple Who's On First records.")

	user_agent := fs.String("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:10.0) Gecko/20100101 Firefox/10.0", "An optional user-agent to append to the -whosonfirst-reader-uri fs")

	var str_properties multi.KeyValueString
	fs.Var(&str_properties, "string-property", "One or more {KEY}={VALUE} fss where {KEY} is a valid tidwall/gjson path and {VALUE} is a string value.")

	var int_properties multi.KeyValueInt64
	fs.Var(&int_properties, "int-property", "One or more {KEY}={VALUE} flags where {KEY} is a valid tidwall/gjson path and {VALUE} is a int(64) value.")

	// var float_properties multi.KeyValueFloat64
	// flag.Var(&float_properties, "float-property", "One or more {KEY}={VALUE} flags where {KEY} is a valid tidwall/gjson path and {VALUE} is a float(64) value.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Fetch one or more Who's on First records and, optionally, their ancestors.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [options] [path1 path2 ... pathN]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nNotes:\n\n")
		fmt.Fprintf(os.Stderr, wordwrap.WrapString("pathN may be any valid Who's On First ID or URI that can be parsed by the go-whosonfirst-uri package.\n\n", 80))
	}

	mode := fs.String("mode", "cli", "Valid options are: cli, lambda")

	flagset.Parse(fs)

	logger := log.Default()
	
	err := flagset.SetFlagsFromEnvVars(fs, "SFOMUSEUM")

	if err != nil {
		log.Fatalf("Failed to set flags from environment variables, %w", err)
	}

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

	*data_writer_uri, err = gh_writer.EnsureGitHubAccessToken(ctx, *data_writer_uri, *token_uri)

	if err != nil {
		log.Fatalf("Failed to append token to data writer, %v", err)
	}

	*properties_writer_uri, err = gh_writer.EnsureGitHubAccessToken(ctx, *properties_writer_uri, *token_uri)

	if err != nil {
		log.Fatalf("Failed to append token to properties writer, %v", err)
	}

	data_wr, err := writer.NewWriter(ctx, *data_writer_uri)

	if err != nil {
		log.Fatalf("Failed to create new data writer, %v", err)
	}

	props_wr, err := writer.NewWriter(ctx, *properties_writer_uri)

	if err != nil {
		log.Fatalf("Failed to create new properties writer, %v", err)
	}

	data_wr.SetLogger(ctx, logger)	
	props_wr.SetLogger(ctx, logger)
	
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

	custom_props := make(map[string]interface{})

	for _, p := range str_properties {
		path := p.Key()
		value := p.Value()

		path = strings.Replace(path, "properties.", "", 1)
		custom_props[path] = value
	}

	for _, p := range int_properties {
		path := p.Key()
		value := p.Value()

		path = strings.Replace(path, "properties.", "", 1)
		custom_props[path] = value
	}

	has_custom := false

	for range custom_props {
		has_custom = true
		break
	}

	if has_custom {
		sfom_opts.CustomProperties = custom_props
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

	// START OF local func to wrap
	// making sure writers are Close()-ed

	import_ids := func(ctx context.Context, ids ...int64) error {

		err := wof_import.ImportFeatures(ctx, import_opts, ids...)

		if err != nil {
			return fmt.Errorf("Failed to import IDs, %v", err)
		}

		err = data_wr.Close(ctx)

		if err != nil {
			return fmt.Errorf("Failed to close data writer, %w", err)
		}

		err = props_wr.Close(ctx)

		if err != nil {
			return fmt.Errorf("Failed to close properties writer, %w", err)
		}

		return nil
	}

	// END OF local func to wrap

	switch *mode {
	case "cli":

		uris := fs.Args()
		count := len(uris)

		feature_ids := make([]int64, count)

		for idx, raw := range uris {

			id, _, err := uri.ParseURI(raw)

			if err != nil {
				log.Fatalf("Unable to parse URI '%s', %v", raw, err)
			}

			feature_ids[idx] = id
		}

		err := import_ids(ctx, feature_ids...)

		if err != nil {
			log.Fatalf("Failed to import IDs, %v", err)
		}

	case "lambda":

		type ImportEvent struct {
			Ids []int64 `json:"ids"`
		}

		handler := func(ctx context.Context, ev ImportEvent) error {

			err := import_ids(ctx, ev.Ids...)

			if err != nil {
				return fmt.Errorf("Failed to import IDs, %v", err)
			}

			return nil
		}

		lambda.Start(handler)

	default:
		log.Fatalf("Invalid or unsupported mode: %s", *mode)
	}

}
