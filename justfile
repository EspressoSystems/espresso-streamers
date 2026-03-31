default: check

build:
    go build ./...

fmt:
    gofmt -w .

fmt-check:
    @test -z "$(gofmt -l .)" || (echo "The following files need formatting:" && gofmt -l . && exit 1)

lint:
    golangci-lint run ./...

test:
    go test -race -count=1 -v ./...

vet:
    go vet ./...

gen-bindings:
    ./scripts/gen_bindings.sh abi/BatchAuthenticator.json > op/bindings/batch_authenticator.go

check: fmt-check vet lint test
