package importer

import (
	"context"
)

const NULL_SCHEME string = "null"

type NullImporter struct {
	Importer
}

func init() {
	ctx := context.Background()
	RegisterImporter(ctx, NULL_SCHEME, NewNullImporter)
}

func NewNullImporter(ctx context.Context, uri string) (Importer, error) {

	i := &NullImporter{}
	return i, nil
}

func (i *NullImporter) ImportIDs(ctx context.Context, ids ...int64) error {
	return nil
}
