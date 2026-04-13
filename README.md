# espresso-streamers

Go library for streaming batches from [Espresso](https://github.com/EspressoSystems/espresso-network) into [Optimism](https://github.com/ethereum-optimism/optimism) L2 nodes. It provides an `EspressoStreamer` interface that handles batch fetching, buffering, and synchronization against Ethereum L1.

## Requirements

- [Go](https://go.dev/) 1.25+
- [just](https://just.systems/)
- [golangci-lint](https://golangci-lint.run/)
- [abigen](https://geth.ethereum.org/docs/tools/abigen) (for generating Go bindings)

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
| `just gen-bindings` | Regenerate Go bindings from ABI |

### Running tests

```bash
just test
```


## License
Copyright
(c) 2022 Espresso Systems espresso-network was developed by Espresso Systems. While we plan to adopt an open source license, we have not yet selected one. As such, all rights are reserved for the time being. Please reach out to us if you have thoughts on licensing.

## Disclaimer
DISCLAIMER: This software is provided "as is" and its security has not been externally audited. Use at your own risk.

DISCLAIMER: The Rust library crates provided in this repository are intended primarily for use by the binary targets in this repository. We make no guarantees of public API stability. If you are building on these crates, reach out by opening an issue to discuss the APIs you need.