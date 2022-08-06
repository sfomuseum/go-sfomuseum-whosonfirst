cli:
	go build -mod vendor -o bin/import-feature cmd/import-feature/main.go
	go build -mod vendor -o bin/refresh-features cmd/refresh-features/main.go
	go build -mod vendor -o bin/ensure-properties cmd/ensure-properties/main.go
	go build -mod vendor -o bin/merge-properties cmd/merge-properties/main.go
	go build -mod vendor -o bin/ensure-features cmd/ensure-features/main.go
