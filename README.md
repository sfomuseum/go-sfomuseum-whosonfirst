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
> ./bin/import-feature -h
Fetch one or more Who's on First records and, optionally, their ancestors.

Usage:
  ./bin/import-feature [options] [path1 path2 ... pathN]

Options:
  -access-token-uri string
    	A valid GitHub API access token. This will be used to replace the "{access_token}" string template in any of the "*-reader-uri" or "*-writer-uri" flag values.
  -data-reader-uri string
    	A valid whosonfirst/go-reader URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/data")
  -data-writer-uri string
    	A valid whosonfirst/go-writer URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/data")
  -enable-filtering
    	If true only source IDs with a lastmodified date greater than their target counterparts will be imported. (default true)
  -filter-reader-uri string
    	A valid whosonfirst/go-reader URI. If empty the value of the -data-reader-uri flag will be used.
  -int-property value
    	One or more {KEY}={VALUE} flags where {KEY} is a valid tidwall/gjson path and {VALUE} is a int(64) value.
  -max-clients int
    	The maximum number of concurrent requests for multiple Who's On First records. (default 10)
  -mode string
    	Valid options are: cli, lambda (default "cli")
  -properties-reader-uri string
    	A valid whosonfirst/go-reader URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties")
  -properties-writer-uri string
    	A valid whosonfirst/go-writer URI. (default "fs:///usr/local/data/sfomuseum-data-whosonfirst/properties")
  -retries int
    	The maximum number of attempts to try fetching a record. (default 3)
  -string-property value
    	One or more {KEY}={VALUE} fss where {KEY} is a valid tidwall/gjson path and {VALUE} is a string value.
  -user-agent string
    	An optional user-agent to append to the -whosonfirst-reader-uri fs (default "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:10.0) Gecko/20100101 Firefox/10.0")
  -whosonfirst-reader-uri string
    	A valid whosonfirst/go-reader URI. (default "https://data.whosonfirst.org/")

Notes:

pathN may be any valid Who's On First ID or URI that can be parsed by the
go-whosonfirst-uri package.
```
#### AWS Lambda

The `import-feature` tool can be compiled and invoked as a Lambda function.

```
$> make lambda-import
if test -f main; then rm -f main; fi
if test -f import-feature.zip; then rm -f import-feature.zip; fi
GOOS=linux go build -mod vendor -o main cmd/import-feature/main.go
zip import-feature.zip main
  adding: main (deflated 58%)
rm -f main
```

##### Environment variables

Once uploaded the Lambda function requires the following environment variables. The environment variables are used to assign their command-line flag equivalents. To determine the environment variable name for a command line flag apply the following rules:

* Upper-case the name of the command line flag
* Replace all instances of "-" with "_" in the command line flag
* Append "SFOMUSEUM_" to the new string

For example the `access-token-uri` flag becomes the `SFOMUSEUM_ACCESS_TOKEN_URI` environment variable.

| Name | Value | Notes |
| --- | --- | --- |
| SFOMUSEUM_ACCESS_TOKEN_URI | awsparamstore://(KEY)?(REGION)&credentials=iam: | A valid `gocloud.dev/runtimevar` URI referencing a valid GitHub API access token. | 
| SFOMUSEUM_DATA_READER_URI | githubapi://(GITHUB_ORG)/(GITHUB_REPO)?access_token={access_token}&prefix=data&branch={data_branch} | If the value of the `?branch=` paramter is "{data_branch}" it will replaced by "(UNIX_TIMESTAMP)-(PROCESS_ID)-data". If the value of `?access_token=` parameter is "{access_token}" it will be replaced by the value derived from the `SFOMUSEUM_ACCESS_TOKEN_URI` environment variable. |
| SFOMUSEUM_DATA_WRITER_URI | githubapi-branch://(GITHUB_ORG)/(GITHUB_REPO)?prefix=data&access_token={access_token}&email=(EMAIL)&description=update%20features&to-branch={data_branch}&merge=true&remove-on-merge=true | If the value of the `?branch=` paramter is "{data_branch}" it will replaced by "(UNIX_TIMESTAMP)-(PROCESS_ID)-data". If the value of `?access_token=` parameter is "{access_token}" it will be replaced by the value derived from the `SFOMUSEUM_ACCESS_TOKEN_URI` environment variable. |
| SFOMUSEUM_ENABLE_FILTERING | true | |
| SFOMUSEUM_FILTER_READER_URI | githubapi://(GITHUB_ORG)/(GITHUB_REPO)?access_token={access_token}&prefix=data | If empty the value of the `SFOMUSEUM_DATA_READER_URI` environment variable will be used. If the value of `?access_token=` parameter is "{access_token}" it will be replaced by the value derived from the `SFOMUSEUM_ACCESS_TOKEN_URI` environment variable. |
| SFOMUSEUM_MODE | lambda | |
| SFOMUSEUM_PROPERTIES_READER_URI | githubapi://(GITHUB_ORG)/(GITHUB_REPO)?access_token={access_token}&prefix=properties&branch={props_branch} | If the value of the `?branch=` paramter is "{props_branch}" it will replaced by "(UNIX_TIMESTAMP)-(PROCESS_ID)-props". If the value of `?access_token=` parameter is "{access_token}" it will be replaced by the value derived from the `SFOMUSEUM_ACCESS_TOKEN_URI` environment variable. |
| SFOMUSEUM_PROPERTIES_WRITER_URI | githubapi-branch://(GITHUB_ORG)/(GITHUB_REPO)?prefix=properties&access_token={access_token}&email=(EMAIL)&description=update%20properties&to-branch={props_branch}&merge=true&remove-on-merge=true | If the value of the `?branch=` paramter is "{props_branch}" it will replaced by "(UNIX_TIMESTAMP)-(PROCESS_ID)-props". If the value of `?access_token=` parameter is "{access_token}" it will be replaced by the value derived from the `SFOMUSEUM_ACCESS_TOKEN_URI` environment variable. |

SFO Museum uses the `githubapi-branch://` writer in order to publish multiple updates to a branch and then merge those updates in the main branch in order to limit the number of webhooks triggered by commits to the main branch. This example reflect's SFO Museum's usage. Any implementation of the [whosonfirst/go-writer/v2.Writer](https://github.com/whosonfirst/go-writer) interface exported by the [whosonfirst/go-writer](https://github.com/whosonfirst/go-writer) and [whosonfirst/go-writer-github](https://github.com/whosonfirst/go-writer-github) can be used.

##### Import Events

The `import-feature` tool, when invoked as a Lambda function, expects to read the IDs to import from a JSON-encoded `ImportEvent` data structure.

```
type ImportEvent struct {
	Ids []int64 `json:"ids"`
}
```

For example:

```
{
  "ids": [
    101714471
  ]
}
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
