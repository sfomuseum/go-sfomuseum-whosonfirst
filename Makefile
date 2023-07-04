GOMOD=vendor

cli:
	go build -mod $(GOMOD) -ldflags="-s -w"  -o bin/import-feature cmd/import-feature/main.go
	go build -mod $(GOMOD) -ldflags="-s -w"  -o bin/refresh-features cmd/refresh-features/main.go
	go build -mod $(GOMOD) -ldflags="-s -w"  -o bin/ensure-properties cmd/ensure-properties/main.go
	go build -mod $(GOMOD) -ldflags="-s -w"  -o bin/merge-properties cmd/merge-properties/main.go
	go build -mod $(GOMOD) -ldflags="-s -w"  -o bin/ensure-features cmd/ensure-features/main.go

lambda:
	@make lambda-import

lambda-import:
	if test -f main; then rm -f main; fi
	if test -f import-feature.zip; then rm -f import-feature.zip; fi
	GOOS=linux go build -mod $(GOMOD) -ldflags="-s -w"  -o main cmd/import-feature/main.go
	zip import-feature.zip main
	rm -f main
