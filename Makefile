cli:
	go build -mod vendor -o bin/import-feature cmd/import-feature/main.go
	go build -mod vendor -o bin/refresh-features cmd/refresh-features/main.go
	go build -mod vendor -o bin/ensure-properties cmd/ensure-properties/main.go
	go build -mod vendor -o bin/merge-properties cmd/merge-properties/main.go
	go build -mod vendor -o bin/ensure-features cmd/ensure-features/main.go

lambda:
	@make lambda-import

lambda-import:
	if test -f main; then rm -f main; fi
	if test -f import-feature.zip; then rm -f import-feature.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/import-feature/main.go
	zip import-feature.zip main
	rm -f main
