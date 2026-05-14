package nitro

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

type AddressValidRangeConfig struct {
	Address string `koanf:"address"`
	From    uint64 `koanf:"from"`
	To      uint64 `koanf:"to"`
}

type AddressValidRange struct {
	Address common.Address `koanf:"address"`
	from    uint64         `koanf:"from"`
	to      uint64         `koanf:"to"`
}

type BatcherAddrMonitor struct {
	addressValidRanges []AddressValidRange
}

func NewBatcherAddressMonitor(addressValidRanges []AddressValidRangeConfig) *BatcherAddrMonitor {
	converted := make([]AddressValidRange, 0, len(addressValidRanges))
	for _, cfg := range addressValidRanges {
		converted = append(converted, AddressValidRange{
			Address: common.HexToAddress(cfg.Address),
			from:    cfg.From,
			to:      cfg.To,
		})
	}
	return &BatcherAddrMonitor{
		addressValidRanges: converted,
	}
}

func (b *BatcherAddrMonitor) IsValid(batcherAddress common.Address, l1Height uint64) bool {
	for _, addr := range b.addressValidRanges {
		if addr.Address == batcherAddress {
			return l1Height >= addr.from && l1Height <= addr.to
		}
	}
	return false
}

func (b *BatcherAddrMonitor) Start(ctx context.Context) error {
	return nil
}
