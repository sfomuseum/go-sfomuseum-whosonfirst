package fetch

import (
	"bytes"
	"context"
	"fmt"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer"
	"io"
	// golog "log"
	"strings"
	"sync"
	"time"
)

type Options struct {
	Timings    bool
	MaxClients int
	Logger     *log.WOFLogger
	Retries    int
}

func DefaultOptions() (*Options, error) {

	logger := log.SimpleWOFLogger()

	o := Options{
		Timings:    false,
		MaxClients: 10,
		Logger:     logger,
		Retries:    0,
	}

	return &o, nil
}

type Fetcher struct {
	reader     reader.Reader
	writer     writer.Writer
	processing *sync.Map
	processed  *sync.Map
	throttle   chan bool
	options    *Options
}

func NewFetcher(ctx context.Context, rdr reader.Reader, wr writer.Writer, opts *Options) (*Fetcher, error) {

	processing := new(sync.Map)
	processed := new(sync.Map)

	max_fetch := opts.MaxClients
	throttle := make(chan bool, max_fetch)

	for i := 0; i < max_fetch; i++ {
		throttle <- true
	}

	f := Fetcher{
		reader:     rdr,
		writer:     wr,
		options:    opts,
		processing: processing,
		processed:  processed,
		throttle:   throttle,
	}

	return &f, nil
}

func (f *Fetcher) FetchIDs(ctx context.Context, ids []int64, belongs_to ...string) ([]int64, error) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	done_ch := make(chan bool)
	err_ch := make(chan error)

	for _, id := range ids {
		go f.FetchID(ctx, id, belongs_to, done_ch, err_ch)
	}

	remaining := len(ids)

	for remaining > 0 {

		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			return nil, fmt.Errorf("Failed to fetch ID (%d remaining), %w", remaining, err)
		default:
			//
		}
	}

	processed := make([]int64, 0)

	f.processed.Range(func(k interface{}, v interface{}) bool {
		id := k.(int64)
		processed = append(processed, id)
		return true
	})

	return processed, nil
}

func (f *Fetcher) FetchID(ctx context.Context, id int64, fetch_belongsto []string, done_ch chan bool, err_ch chan error) {

	defer func() {
		done_ch <- true
	}()

	select {

	case <-ctx.Done():
		return
	default:
		// pass
	}

	err := f.fetchID(ctx, id, fetch_belongsto...)

	if err != nil {
		err_ch <- err
	}
}

func (f *Fetcher) fetchID(ctx context.Context, id int64, belongs_to ...string) error {

	if id < 0 {
		return nil
	}

	_, ok := f.processed.Load(id)

	if ok {
		f.options.Logger.Status("%d has already been processed, skipping", id)
		return nil
	}

	_, ok = f.processing.LoadOrStore(id, true)

	if ok {
		f.options.Logger.Debug("%d is being processed, skipping", id)
		return nil
	}

	if f.options.Timings {

		t1 := time.Now()

		defer func() {
			f.options.Logger.Status("Time to process %d: %v", id, time.Since(t1))
		}()
	}

	<-f.throttle

	f.options.Logger.Debug("processing (%d)", id)

	defer func() {
		f.throttle <- true
		f.processing.Delete(id)
	}()

	path, err := uri.Id2RelPath(id)

	if err != nil {
		return err
	}

	var infile io.ReadCloser
	var read_err error

	attempts := f.options.Retries + 1

	for attempts > 0 {

		infile, read_err = f.reader.Read(ctx, path)

		attempts = attempts - 1

		if read_err == nil {
			break
		}

		//logger.Warning("Failed to fetch %d because %s (remaining attempts: %d)", wofid, err, attempts)
	}

	if read_err != nil {
		return read_err
	}

	defer func() {
		infile.Close()
	}()

	body, err := io.ReadAll(infile)

	if err != nil {
		return err
	}

	br := bytes.NewReader(body)
	fh, err := ioutil.NewReadSeekCloser(br)

	if err != nil {
		return err
	}

	_, write_err := f.writer.Write(ctx, path, fh)

	if write_err != nil {
		return write_err
	}

	f.processed.Store(id, true)

	count_belongs_to := len(belongs_to)

	if count_belongs_to > 0 {

		br := bytes.NewReader(body)
		fh := io.NopCloser(br)

		ft, err := feature.LoadWOFFeatureFromReader(fh)

		if err != nil {
			return err
		}

		ids := make([]int64, 0)

		if count_belongs_to == 1 && belongs_to[0] == "all" {

			ids = whosonfirst.BelongsTo(ft)

		} else {

			hiers := whosonfirst.Hierarchies(ft)

			for _, h := range hiers {

				for pt, other_id := range h {

					possible := true

					for _, candidate_id := range ids {

						if other_id == candidate_id {
							possible = false
							break
						}
					}

					if possible == false {
						continue
					}

					pt = strings.Replace(pt, "_id", "", -1)

					for _, candidate_pt := range belongs_to {

						if pt == candidate_pt {
							ids = append(ids, other_id)
							break
						}
					}

				}
			}
		}

		if len(ids) > 0 {

			_, err = f.FetchIDs(ctx, ids)

			if err != nil {
				// return err
			}
		}
	}

	return nil
}
