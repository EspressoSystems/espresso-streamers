# Default recipe: run all checks
default: check

# Format all Go source files
fmt:
    gofmt -w .

# Check formatting (fails if files are unformatted)
fmt-check:
    @test -z "$(gofmt -l .)" || (echo "The following files need formatting:" && gofmt -l . && exit 1)

# Run linter
lint:
    golangci-lint run ./...

# Run all tests
test:
    go test -race -count=1 ./...

# Run go vet
vet:
    go vet ./...

# Run all checks (fmt, vet, lint, test)
check: fmt-check vet lint test
