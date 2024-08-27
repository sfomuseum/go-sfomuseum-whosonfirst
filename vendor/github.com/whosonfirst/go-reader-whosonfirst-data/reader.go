package reader

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"sync"
	"time"

	_ "github.com/whosonfirst/go-reader-github"
	_ "github.com/whosonfirst/go-reader-http"

	wof_reader "github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/resolver"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

type WhosOnFirstDataReader struct {
	wof_reader.Reader
	throttle     <-chan time.Time
	provider     string
	organization string
	repo         string
	branch       string
	prefix       string
	repos        *sync.Map
	readers      *sync.Map
	resolver     resolver.Resolver
}

func init() {

	ctx := context.Background()
	err := wof_reader.RegisterReader(ctx, "whosonfirst-data", NewWhosOnFirstDataReader)

	if err != nil {
		panic(err)
	}
}

func NewWhosOnFirstDataReader(ctx context.Context, uri string) (wof_reader.Reader, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse reader URI '%s', %w", uri, err)
	}

	q := u.Query()

	provider := q.Get("provider")
	org := q.Get("organization")
	repo := q.Get("repo")
	branch := q.Get("branch")
	prefix := q.Get("prefix")

	if provider == "" {
		provider = "github"
	}

	if org == "" {
		org = "whosonfirst-data"
	}

	if prefix == "" {
		prefix = "data"
	}

	// This is a specific whosonfirst-data -ism
	// https://github.com/whosonfirst-data/whosonfirst-data/issues/1919

	if branch == "" && org == "whosonfirst-data" {
		branch = "master"
	}

	fa_uri := q.Get("findingaid-uri")

	if fa_uri == "" {

		// START OF cue the Inception Horn

		// See this? We're querying the data.whosonfirst.org S3 bucket
		// of data for a record in order to glean the repo for that
		// record to fetch the record (again) from GitHub. This is necessary
		// in advance of setting up a KV finding aid lookup for all of
		// Who's On First (20220510/thisisaaronland)

		r, err := url.Parse("https://data.whosonfirst.org")

		if err != nil {
			return nil, fmt.Errorf("Failed to parse URL, %w", err)
		}

		rq := r.Query()
		rq.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:10.0) Gecko/20100101 Firefox/10.0")

		r.RawQuery = rq.Encode()
		reader_uri := r.String()

		f := new(url.URL)
		f.Scheme = "reader"

		fq := f.Query()
		fq.Set("reader", reader_uri)
		fq.Set("strategy", "uri")

		f.RawQuery = fq.Encode()
		fa_uri = f.String()

		// END OF cue the Inception Horn
	}

	rslvr, err := resolver.NewResolver(ctx, fa_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create finding aid resolver for '%s', %w", fa_uri, err)
	}

	rate := time.Second / 3
	throttle := time.Tick(rate)

	repos := new(sync.Map)
	readers := new(sync.Map)

	r := &WhosOnFirstDataReader{
		throttle:     throttle,
		provider:     provider,
		organization: org,
		repo:         repo,
		branch:       branch,
		repos:        repos,
		prefix:       prefix,
		readers:      readers,
		resolver:     rslvr,
	}

	return r, nil
}

func (r *WhosOnFirstDataReader) Read(ctx context.Context, uri string) (io.ReadSeekCloser, error) {

	<-r.throttle

	select {
	case <-ctx.Done():
		return nil, nil
	default:
		// pass
	}

	gh_r, err := r.getReader(ctx, uri)

	if err != nil {
		return nil, err
	}

	slog.Debug("Read URI", "reader", fmt.Sprintf("%T", gh_r), "uri", uri)
	return gh_r.Read(ctx, uri)
}

func (r *WhosOnFirstDataReader) ReaderURI(ctx context.Context, uri string) string {

	gh_r, err := r.getReader(ctx, uri)

	if err != nil {
		return "" // nil, fmt.Errorf("Failed to create reader for '%s' (%s), %w", uri, repo_name, err)
	}

	return gh_r.ReaderURI(ctx, uri)
}

func (r *WhosOnFirstDataReader) getReader(ctx context.Context, uri string) (wof_reader.Reader, error) {

	repo_name := r.repo

	if repo_name == "" {

		this_repo, err := r.getRepo(ctx, uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to determine repo for '%s', %w", uri, err)
		}

		repo_name = this_repo
	}

	gh_r, err := r.getReaderWithRepo(ctx, repo_name)

	if err != nil {
		return nil, fmt.Errorf("Failed to create reader for '%s' (%s), %w", uri, repo_name, err)
	}

	return gh_r, nil
}

func (r *WhosOnFirstDataReader) getReaderWithRepo(ctx context.Context, repo string) (wof_reader.Reader, error) {

	logger := slog.Default()
	logger = logger.With("repo", repo)

	v, ok := r.readers.Load(repo)

	if ok {
		gh_r := v.(wof_reader.Reader)
		return gh_r, nil
	}

	gh_q := url.Values{}

	if r.branch != "" {
		gh_q.Set("branch", r.branch)
	}

	if r.prefix != "" {
		gh_q.Set("prefix", r.prefix)
	}

	gh_uri := url.URL{}
	gh_uri.Scheme = r.provider
	gh_uri.Host = r.organization
	gh_uri.Path = repo
	gh_uri.RawQuery = gh_q.Encode()

	reader_uri := gh_uri.String()

	logger.Debug("Create new reader", "reader_uri", reader_uri)

	gh_r, err := wof_reader.NewReader(ctx, reader_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create reader for '%s', %w", reader_uri, err)
	}

	go func() {
		r.readers.Store(repo, gh_r)
	}()

	return gh_r, nil
}

func (r *WhosOnFirstDataReader) getRepo(ctx context.Context, path string) (string, error) {

	logger := slog.Default()
	logger = logger.With("path", path)

	v, ok := r.repos.Load(path)

	if ok {
		repo_name := v.(string)
		logger.Debug("Return repo from cache", "repo", repo_name)
		return repo_name, nil
	}

	id, _, err := uri.ParseURI(path)

	if err != nil {
		logger.Debug("Failed to parse path", "error", err)
		return "", fmt.Errorf("Failed to parse %s, %w", path, err)
	}

	repo_name, err := r.resolver.GetRepo(ctx, id)

	if err != nil {
		logger.Debug("Failed to resolve ID", "id", id, "error", err)
		return "", fmt.Errorf("Failed to resolve repo name for '%s', %w", path, err)
	}

	go func() {
		r.repos.Store(path, repo_name)
	}()

	logger.Debug("Resolve repo", "repo", repo_name)
	return repo_name, nil
}
