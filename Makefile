GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")

LDFLAGS=-s -w

cli:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)"  -o bin/import-feature cmd/import-feature/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)"  -o bin/refresh-features cmd/refresh-features/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)"  -o bin/ensure-properties cmd/ensure-properties/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)"  -o bin/merge-properties cmd/merge-properties/main.go
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)"  -o bin/ensure-features cmd/ensure-features/main.go

lambda:
	@make lambda-import

lambda-import:
	if test -f bootstrap; then rm -f bootstrap; fi
	if test -f import-feature.zip; then rm -f import-feature.zip; fi
	GOARCH=arm64 GOOS=linux go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -tags lambda.norpc -o bootstrap cmd/import-feature/main.go
	zip import-feature.zip bootstrap
	rm -f bootstrap
