# espresso-streamers

Go library for streaming batches from [Espresso](https://github.com/EspressoSystems/espresso-network) into [Optimism](https://github.com/ethereum-optimism/optimism) L2 nodes. It provides an `EspressoStreamer` interface that handles batch fetching, buffering, and synchronization against Ethereum L1.

## Requirements

- [Go](https://go.dev/) 1.25+
- [just](https://just.systems/)
- [golangci-lint](https://golangci-lint.run/)

Or with Nix:

```bash
nix develop
```

## Usage

```go
import "github.com/EspressoSystems/espresso-streamers/op"
```

## Development

| Command | Description |
|---|---|
| `just test` | Run tests |
| `just vet` | Run `go vet` |
| `just lint` | Run `golangci-lint` |
| `just fmt` | Format code |
| `just check` | Run all of the above |

### Running tests

```bash
just test
```
