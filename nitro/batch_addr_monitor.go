package nitro

import (
	"github.com/ethereum/go-ethereum/common"
)

type AddressValidRangeConfig struct {
	Address string `koanf:"address"`
	From    uint64 `koanf:"from"`
	To      uint64 `koanf:"to"`
}

type AddressValidRange struct {
	address common.Address
	from    uint64
	to      uint64
}

func (a AddressValidRange) isValidAt(address common.Address, l1Height uint64) bool {
	return a.address == address && l1Height >= a.from && l1Height <= a.to
}

type BatcherAddrMonitor struct {
	addressValidRanges []AddressValidRange
}

func NewBatcherAddressMonitor(addressValidRanges []AddressValidRangeConfig) *BatcherAddrMonitor {
	converted := make([]AddressValidRange, 0, len(addressValidRanges))
	for _, cfg := range addressValidRanges {
		converted = append(converted, AddressValidRange{
			address: common.HexToAddress(cfg.Address),
			from:    cfg.From,
			to:      cfg.To,
		})
	}
	return &BatcherAddrMonitor{
		addressValidRanges: converted,
	}
}

// AddressValidator defines an interface for verifying if a given address is
// valid at a specific L1 block height.
type AddressValidator interface {
	// IsValid determines whether a given Address is a valid signer for data
	// with a corresponding L1 Block Height.
	IsValid(address common.Address, l1Height uint64) bool
}

func (b *BatcherAddrMonitor) IsValid(batcherAddress common.Address, l1Height uint64) bool {
	for _, addr := range b.addressValidRanges {
		if addr.isValidAt(batcherAddress, l1Height) {
			return true
		}
	}
	return false
}
