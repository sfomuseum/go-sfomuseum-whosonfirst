# go-sfomuseum-whosonfirst

Go package for working with whosonfirst-data repositories in a SFO Museum context.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/sfomuseum/go-sfomuseum-whosonfirst.svg)](https://pkg.go.dev/github.com/sfomuseum/go-sfomuseum-whosonfirst)

## Tools

```
$> make cli
go build -mod vendor -o bin/import-feature cmd/import-feature/main.go
go build -mod vendor -o bin/refresh-features cmd/refresh-features/main.go
go build -mod vendor -o bin/ensure-properties cmd/ensure-properties/main.go
go build -mod vendor -o bin/merge-properties cmd/merge-properties/main.go
```

### ensure-properties

```
$> ./bin/ensure-properties -h
ensure-properties iterates over a collection of Who's On First records and ensures that there is a corresponding properties JSON file.
Usage:
	 ./bin/ensure-properties [options] uri(N) uri(N)
  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/v2 URI. (default "repo://")
  -properties-reader-uri string
    	A valid whosonfirst/go-reader.Reader URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties")
  -properties-writer-uri string
    	A valid whosonfirst/go-writer.Writer URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties")
```

### import-feature

```
$> ./bin/import-feature -h
Fetch one or more Who's on First records and, optionally, their ancestors.

Usage:
  ./bin/import-feature [options] [path1 path2 ... pathN]

Options:
  -data-reader-uri string
    	A valid whosonfirst/go-reader URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/data")
  -data-writer-uri string
    	A valid whosonfirst/go-writer URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/data")
  -int-property value
    	One or more {KEY}={VALUE} flags where {KEY} is a valid tidwall/gjson path and {VALUE} is a int(64) value.
  -max-clients int
    	The maximum number of concurrent requests for multiple Who's On First records. (default 10)
  -properties-reader-uri string
    	A valid whosonfirst/go-reader URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties")
  -properties-writer-uri string
    	A valid whosonfirst/go-writer URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties")
  -retries int
    	The maximum number of attempts to try fetching a record. (default 3)
  -string-property value
    	One or more {KEY}={VALUE} flags where {KEY} is a valid tidwall/gjson path and {VALUE} is a string value.
  -user-agent string
    	An optional user-agent to append to the -whosonfirst-reader-uri flag (default "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:10.0) Gecko/20100101 Firefox/10.0")
  -whosonfirst-reader-uri string
    	A valid whosonfirst/go-reader URI. (default "https://data.whosonfirst.org/")

Notes:

pathN may be any valid Who's On First ID or URI that can be parsed by the go-whosonfirst-uri package.
```

### merge-properties

```
$> ./bin/merge-properties -h
merge-properties iterates over a collection of Who's On First records and merges custom properties.
Usage:
	 ./bin/merge-properties [options] record(N) record(N)
  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/v2 URI. (default "repo://")
  -properties-reader-uri string
    	A valid whosonfirst/go-reader.Reader URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties")
  -reader-uri string
    	A valid whosonfirst/go-reader.Reader URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/data")
  -writer-uri string
    	A valid whosonfirst/go-writer.Writer URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/data")
```

### refresh-feature

```
$> ./bin/refresh-features -h
refresh-features is a command line tool for refreshing all the source Who's On First records in the sfomuseum-data-whosonfirst repository.
Usage:
	 ./bin/refresh-features [options]
  -data-iterator-source string
    	... (default "/usr/local/data/sfomuseum-data-whosonfirst")
  -data-iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/v2 URI (default "repo://")
  -data-reader-uri string
    	A valid whosonfirst/go-reader URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/data")
  -data-writer-uri string
    	A valid whosonfirst/go-writer URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/data")
  -max-clients int
    	The maximum number of concurrent requests for multiple Who's On First records. (default 10)
  -properties-reader-uri string
    	A valid whosonfirst/go-reader URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties")
  -properties-writer-uri string
    	A valid whosonfirst/go-writer URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties")
  -retries int
    	The maximum number of attempts to try fetching a record. (default 3)
  -user-agent string
    	An optional user-agent to append to the -whosonfirst-reader-uri flag (default "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:10.0) Gecko/20100101 Firefox/10.0")
  -whosonfirst-reader-uri string
    	A valid whosonfirst/go-reader URI. (default "whosonfirst-data://")
```

## See also

* https://github.com/sfomuseum-data
* https://github.com/sfomuseum-data/sfomuseum-data-whosonfirst
* https://github.com/whosonfirst-data
