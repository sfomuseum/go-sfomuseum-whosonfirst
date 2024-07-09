package filter

import (
	"context"
	"fmt"

	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
)

// FilterByLastModified() iterates through 'ids' and removes entries that can be read by 'target_r' and who's "wof:lastmodified" value
// is greater than the corresponding record read by `source_r`.
func FilterByLastModified(ctx context.Context, source_r reader.Reader, target_r reader.Reader, ids ...int64) ([]int64, error) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	filtered := make([]int64, 0)

	done_ch := make(chan bool)
	err_ch := make(chan error)
	id_ch := make(chan int64)

	for _, id := range ids {

		go func(ctx context.Context, id int64) {

			defer func() {
				done_ch <- true
			}()

			t, err := wof_reader.LoadBytes(ctx, target_r, id)

			if err != nil {
				id_ch <- id
				return
			}

			s, err := wof_reader.LoadBytes(ctx, source_r, id)

			if err != nil {
				err_ch <- fmt.Errorf("Failed to load %d from source, %w", id, err)
				return
			}

			t_lastmod := properties.LastModified(t)
			s_lastmod := properties.LastModified(s)

			if s_lastmod > t_lastmod {
				id_ch <- id
			}

		}(ctx, id)

	}

	remaining := len(ids)

	for remaining > 0 {

		select {
		case <-ctx.Done():
			break
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			return nil, fmt.Errorf("Failed to filter IDs, %w", err)
		case i := <-id_ch:
			filtered = append(filtered, i)
		}
	}

	return filtered, nil
}
