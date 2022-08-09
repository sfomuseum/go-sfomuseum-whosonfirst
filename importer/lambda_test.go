package importer

import (
	"context"
	"flag"
	"testing"
)

// go test -v -run TestLambdaImporter -importer-uri 'lambda://ImportWhosOnFirstFeature?region={REGION}&credentials={CREDENTIALS}&type=Event' -id 1310327797

var importer_uri = flag.String("importer-uri", "", "")
var id = flag.Int64("id", 0, "")

func TestLambdaImporter(t *testing.T) {

	if *importer_uri == "" {
		t.Skip()
	}

	ctx := context.Background()

	i, err := NewImporter(ctx, *importer_uri)

	if err != nil {
		t.Fatalf("Failed to create importer for %s, %v", *importer_uri, err)
	}

	err = i.ImportIDs(ctx, *id)

	if err != nil {
		t.Fatalf("Failed to import IDs, %v", err)
	}
}
