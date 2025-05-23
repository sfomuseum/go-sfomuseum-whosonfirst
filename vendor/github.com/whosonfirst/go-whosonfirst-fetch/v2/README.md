# go-whosonfirst-fetch

Tools for fetching Who's On First records and their ancestors.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-whosonfirst-fetch.svg)](https://pkg.go.dev/github.com/whosonfirst/go-whosonfirst-fetch)

## Example

```
import (
	"context"

	_ "github.com/whosonfirst/go-reader-whosonfirst-data"
	
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-writer"
	"github.com/whosonfirst/go-whosonfirst-fetch"
)

func main() {

	ctx := context.Background()

	r, _ := reader.NewReader(ctx, "whosonfirst-data://")
	wr, _ := writer.NewWriter(ctx, "stdout://)

	fetcher_opts, _ := fetch.DefaultOptions()
	
	fetcher, _ := fetch.NewFetcher(ctx, r, wr, fetcher_opts)

	ids := []int{
		1360695651,
	}

	belongs_to := []string{
		"all",
	}
	
	fetcher.FetchIDs(ctx, ids, belongs_to...)
}
```

_Error handling omitted for the sake of brevity._

The example above would result in the following Who's On First documents being retrieved:

* [136/069/565/1/1360695651.geojson](https://spelunker.whosonfirst.org/id/1360695651)
* [137/769/436/9/1377694369.geojson](https://spelunker.whosonfirst.org/id/1377694369)
* [102/063/913/102063913.geojson](https://spelunker.whosonfirst.org/id/102063913)
* [102/191/581/102191581.geojson](https://spelunker.whosonfirst.org/id/102191581)
* [101/756/499/101756499.geojson](https://spelunker.whosonfirst.org/id/101756499)
* [856/331/11/85633111.geojson](https://spelunker.whosonfirst.org/id/85633111)
* [856/825/53/85682553.geojson](https://spelunker.whosonfirst.org/id/85682553)

## Tools

To build binary versions of these tools run the `cli` Makefile target. For example:

```
$> make cli
go build -mod vendor -o bin/fetch cmd/fetch/main.go
```

### fetch

```
$> ./bin/fetch -h
Fetch one or more Who's on First records and, optionally, their ancestors.

Usage:
  ./bin/fetch [options] [path1 path2 ... pathN]

Options:
  -belongs-to value
    	One or more placetypes that a given ID may belong to to also fetch. You may also pass 'all' as a short-hand to fetch the entire hierarchy for a place.
  -max-clients int
    	The maximum number of concurrent requests for multiple Who's On First records. (default 10)
  -reader-uri string
    	A valid whosonfirst/go-reader URI. (default "whosonfirst-data://")
  -retries int
    	The maximum number of attempts to try fetching a record. (default 3)
  -writer-uri string
    	A valid whosonfirst/go-writer URI. (default "null://")

Notes:

pathN may be any valid Who's On First ID or URI that can be parsed by the
go-whosonfirst-uri package.
```

Fetch one or more Who's on First records and, optionally, their ancestors. For example:

```
$> ./bin/fetch -belongs-to all -writer-uri stdout:// 1360695651 | grep 'wof:name'
    "wof:name":"Berlin Brandenburg Willy Brandt Airport",
    "wof:name":"Germany",
    "wof:name":"Sch\u00f6nefeld",
    "wof:name":"Europe",
    "wof:name":"Sch\u00f6nefeld",
    "wof:name":"Dahme-Spreewald",
    "wof:name":"Brandenburg",
```

Under the hood this tool uses a number of other packages for handling specific tasks:

* It uses the `go-reader` package to read Who's On First records that are fetched.
* It uses the `go-writer` package to write those records.

### Readers

Readers exported by the following packages are available to the `fetch` tool, by default:

* https://github.com/whosonfirst/go-reader
* https://github.com/whosonfirst/go-reader-whosonfirst-data
* https://github.com/whosonfirst/go-reader-http

### Writers

Writers exported by the following packages are available to the `fetch` tool:

* https://github.com/whosonfirst/go-writer

### Alternate readers and writers

As mentioned above the tool supports a limited set of default "readers" and "writers". If you need to add additional readers or writers you'll need to clone the code in [cmd/fetch/main.go](cmd/fetch/main.go) and add the relevant `import` statements.

The "guts" of the `fetch` tool live in [app/fetch/fetch.go](app/fetch/fetch.go) which allows `cmd/fetch/main.go` to be very small in order to make add custom imports easy (or at least fast). 

For example, this is what the code for a custom `fetch` tool which could also read from GitHub and write to a GeoParquet database would look like:

```
package main

import (
	"context"
	"log"

	_ "github.com/whosonfirst/go-reader-http"
	_ "github.com/whosonfirst/go-reader-github"
	_ "github.com/whosonfirst/go-reader-whosonfirst-data"
	_ "github.com/whosonfirst/go-writer-geoparquet/v3"
	
	"github.com/whosonfirst/go-whosonfirst-fetch/app/fetch"
)

func main() {

	ctx := context.Background()
	err := fetch.Run(ctx)

	if err != nil {
		log.Fatalf("Failed to fetch records, %v", err)
	}
}
```

## See also

* https://github.com/whosonfirst/go-reader
* https://github.com/whosonfirst/go-reader-whosonfirst-data
* https://github.com/whosonfirst/go-reader-http
* https://github.com/whosonfirst/go-writer
* https://github.com/whosonfirst/go-whosonfirst-findingaid