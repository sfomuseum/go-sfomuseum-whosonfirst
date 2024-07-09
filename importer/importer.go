package importer

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
)

type Importer interface {
	ImportIDs(context.Context, ...int64) error
}

var importer_roster roster.Roster

// ImporterInitializationFunc is a function defined by individual importer package and used to create
// an instance of that importer
type ImporterInitializationFunc func(ctx context.Context, uri string) (Importer, error)

// RegisterImporter registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Importer` instances by the `NewImporter` method.
func RegisterImporter(ctx context.Context, scheme string, init_func ImporterInitializationFunc) error {

	err := ensureImporterRoster()

	if err != nil {
		return err
	}

	return importer_roster.Register(ctx, scheme, init_func)
}

func ensureImporterRoster() error {

	if importer_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		importer_roster = r
	}

	return nil
}

// NewImporter returns a new `Importer` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `ImporterInitializationFunc`
// function used to instantiate the new `Importer`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterImporter` method.
func NewImporter(ctx context.Context, uri string) (Importer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := importer_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(ImporterInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureImporterRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range importer_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}
