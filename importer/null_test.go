package importer

import (
	"context"
	"testing"
)

func TestNullImporter(t *testing.T) {

	ctx := context.Background()

	importer_uri := "null://"

	i, err := NewImporter(ctx, importer_uri)

	if err != nil {
		t.Fatalf("Failed to create importer for %s, %v", importer_uri, err)
	}

	err = i.ImportIDs(ctx, 1310327797)

	if err != nil {
		t.Fatalf("Failed to import IDs, %v", err)
	}
}
