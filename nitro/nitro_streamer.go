package nitro

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	espressoTypes "github.com/EspressoSystems/espresso-network/sdks/go/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

const HOTSHOT_RANGE_LIMIT = 100

var (
	ErrFailedToFetchTransactions = errors.New("failed to fetch transactions")
	ErrPayloadHadNoMessages      = errors.New("ParseHotShotPayload found no messages, the transaction may be empty")
	ErrUserDataHashNot32Bytes    = errors.New("user data hash is not 32 bytes")
)

type EspressoStreamerInterface interface {
	Start(ctx context.Context) error
	// Peek returns the next message in the streamer's buffer. If the message is not
	// in the buffer, it will return nil.
	Peek() *MessageWithMetadataAndPos
	// Advance moves the current message position to the next message.
	Advance()
	// AdvanceTo moves the current message position to the specified message.
	AdvanceTo(toPos uint64)
	// Reset sets the current message position and the next hotshot block number.
	Reset(currentMessagePos uint64, currentHostshotBlock uint64)

	GetCurrentEarliestHotShotBlockNumber(pos uint64) uint64

	StopAndWait()
}

type EspressoClientInterface interface {
	FetchLatestBlockHeight(ctx context.Context) (uint64, error)
	FetchHeaderByHeight(ctx context.Context, blockHeight uint64) (espressoTypes.HeaderImpl, error)
	FetchNamespaceTransactionsInRange(ctx context.Context, from, to uint64, namespace uint64) ([]espressoTypes.NamespaceTransactionsRangeData, error)
}

type EspressoStreamer struct {
	espressoClient            EspressoClientInterface
	nextHotshotBlockNum       uint64
	currentMessagePos         uint64
	namespace                 uint64
	messageWithMetadataAndPos map[uint64]*MessageWithMetadataAndPos
	highestPos                uint64

	messageLock sync.RWMutex
	retryTime   time.Duration

	monitor *BatcherAddrMonitor

	log    log.Logger
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

var _ EspressoStreamerInterface = (*EspressoStreamer)(nil)

func NewEspressoStreamer(
	namespace uint64,
	nextHotshotBlockNum uint64,
	espressoClient EspressoClientInterface,
	addressValidRanges []AddressValidRangeConfig,
	retryTime time.Duration,
	startMessagePos uint64,
	logger log.Logger,
) *EspressoStreamer {
	monitor := NewBatcherAddressMonitor(addressValidRanges)
	return &EspressoStreamer{
		espressoClient:            espressoClient,
		nextHotshotBlockNum:       nextHotshotBlockNum,
		namespace:                 namespace,
		retryTime:                 retryTime,
		currentMessagePos:         startMessagePos,
		messageWithMetadataAndPos: make(map[uint64]*MessageWithMetadataAndPos),
		highestPos:                1,
		monitor:                   monitor,
		log:                       logger,
	}
}

func (s *EspressoStreamer) CanBatcherAddressSend(ctx context.Context, address common.Address) (bool, error) {
	latest, err := s.espressoClient.FetchLatestBlockHeight(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to fetch espresso latest block height: %w", err)
	}
	// Even though we can query the latest block height, the node may not yet serve
	// the header at that exact height. Using `latest-1` avoids spurious errors
	// where this function would otherwise always fail. This is safe because
	// Espresso block production is much faster than L1, and the L1 lag has
	// already been accounted for in the batcher address monitor.
	// TODO: Figure out why this doesn't work without `-1`.
	// It might be just a dev node issue.
	header, err := s.espressoClient.FetchHeaderByHeight(ctx, latest-1)
	if err != nil {
		return false, fmt.Errorf("failed to fetch espresso block header: %w", err)
	}
	l1Finalized := header.Header.GetL1Finalized()
	if l1Finalized == nil {
		return false, fmt.Errorf("l1 finalized not found")
	}
	return s.monitor.IsValid(address, l1Finalized.Number), nil
}

// GetMessageCount counts consecutive positions from currentMessagePos.
func (s *EspressoStreamer) GetMessageCount() uint64 {
	s.messageLock.RLock()
	defer s.messageLock.RUnlock()
	count := s.currentMessagePos
	for {
		if _, ok := s.messageWithMetadataAndPos[count]; !ok {
			return count
		}
		count++
	}
}

func (s *EspressoStreamer) Reset(currentMessagePos uint64, currentHotshotBlock uint64) {
	s.messageLock.Lock()
	defer s.messageLock.Unlock()

	hotshotBlockNum := currentHotshotBlock

	s.currentMessagePos = currentMessagePos
	s.nextHotshotBlockNum = hotshotBlockNum
	s.highestPos = currentMessagePos
	s.messageWithMetadataAndPos = make(map[uint64]*MessageWithMetadataAndPos)
}

func (s *EspressoStreamer) Peek() *MessageWithMetadataAndPos {
	s.messageLock.RLock()
	defer s.messageLock.RUnlock()

	return s.messageWithMetadataAndPos[s.currentMessagePos]
}

func (s *EspressoStreamer) GetMsg(pos uint64) *MessageWithMetadataAndPos {
	s.messageLock.RLock()
	defer s.messageLock.RUnlock()

	return s.messageWithMetadataAndPos[pos]
}

func (s *EspressoStreamer) Advance() {
	s.messageLock.Lock()
	defer s.messageLock.Unlock()
	delete(s.messageWithMetadataAndPos, s.currentMessagePos)
	s.currentMessagePos += 1
}

func (s *EspressoStreamer) AdvanceTo(toPos uint64) {
	s.messageLock.Lock()
	defer s.messageLock.Unlock()
	if toPos <= s.currentMessagePos {
		return
	}

	for pos := s.currentMessagePos; pos < toPos; pos++ {
		delete(s.messageWithMetadataAndPos, pos)
	}

	s.currentMessagePos = toPos
}

// QueueMessagesFromHotshot fetches hotshot blocks and parses them.
// parseHotShotPayloadFn is exposed for testing.
func (s *EspressoStreamer) QueueMessagesFromHotshot(
	ctx context.Context,
	parseHotShotPayloadFn func(tx espressoTypes.Bytes, l1Height uint64) error,
) error {
	startHotshotBlockNum := s.nextHotshotBlockNum
	toBlock, err := fetchNextHotshotBlock(
		ctx,
		s.espressoClient,
		startHotshotBlockNum,
		parseHotShotPayloadFn,
		s.namespace,
		s.log,
	)
	if err != nil {
		return err
	}

	s.messageLock.Lock()
	defer s.messageLock.Unlock()
	// Don't jump to toBlock if reset() was called while we were fetching.
	if s.nextHotshotBlockNum == startHotshotBlockNum {
		s.nextHotshotBlockNum = toBlock
	}
	return nil
}

func (s *EspressoStreamer) GetCurrentEarliestHotShotBlockNumber(pos uint64) uint64 {
	s.messageLock.RLock()
	defer s.messageLock.RUnlock()
	if len(s.messageWithMetadataAndPos) == 0 {
		// Streamer is empty; earliest hotshot block is the next one we'll fetch.
		return s.nextHotshotBlockNum
	}
	if msg, exists := s.messageWithMetadataAndPos[pos]; exists {
		return msg.HotshotHeight
	}

	// pos not found — find the minimum height among all buffered positions >= pos.
	minHeight := s.nextHotshotBlockNum
	for nextPos, msg := range s.messageWithMetadataAndPos {
		if nextPos >= pos && msg.HotshotHeight < minHeight {
			minHeight = msg.HotshotHeight
		}
	}
	return minHeight
}

// ChainID returns the expected ChainID for the Chain.
//
// This method serves a couple of purposes.
//
// First is codefies the assumption that the `namespace` is the `ChainID`. If
// this assumption is incorrect, then this method will need to be modified.
//
// Second, it allows for a clean modification of the `ChainID` implementation
// without needing to modify downstream dependencies if it ever does change.
func (s *EspressoStreamer) ChainID() uint64 {
	return s.namespace
}

func (s *EspressoStreamer) parseEspressoTransaction(tx espressoTypes.Bytes, l1Height uint64) error {
	message, err := ParseNitroMessagesFromHotShot(tx)
	if err != nil {
		s.log.Warn("failed to parse hotshot payload", "err", err)
		return fmt.Errorf("failed to parse hotshot payload: %w", err)
	}

	// Testing for no messages
	{
		it := message.MessageIterator()
		if _, ok := it.NextMessage(0); !ok {
			return ErrPayloadHadNoMessages
		}
	}

	// Let's verify the payload signature
	// NOTE: this assumes that the namespace **IS** the Chain ID.  If this is
	// not, then we'll need to introduce a breaking change that determines
	if err := message.VerifySignature(s.monitor, s.ChainID(), l1Height); err != nil {
		s.log.Warn("failed to verify batch poster signature", "err", err)
		return fmt.Errorf("failed to verify batch poster signature: %w", err)
	}

	iterator := message.MessageIterator()

	s.messageLock.Lock()
	defer s.messageLock.Unlock()

	for {
		msg, ok := iterator.NextMessage(s.nextHotshotBlockNum)
		if !ok {
			// We're out of messages
			break
		}

		if msg.Pos < s.currentMessagePos {
			log.Warn("message index is less than current message pos, skipping", "msgPos", msg.Pos, "currentMessagePos", s.currentMessagePos)
			continue
		}
		if _, exists := s.messageWithMetadataAndPos[msg.Pos]; exists {
			log.Warn("duplicate message position, skipping", "msgPos", msg.Pos)
			continue
		}

		if msg.Pos > s.highestPos {
			s.highestPos = msg.Pos
		}

		// Check if we have a higher position in an earlier block.
		currHeight := msg.HotshotHeight
		for nextPos := msg.Pos + 1; nextPos <= s.highestPos; nextPos++ {
			if higherPos, ok := s.messageWithMetadataAndPos[nextPos]; ok && higherPos.HotshotHeight < currHeight {
				currHeight = higherPos.HotshotHeight
				msg.HotshotHeight = currHeight
			}
		}

		s.messageWithMetadataAndPos[msg.Pos] = &msg

		s.log.Debug("Added message to queue", "message", msg.Pos)
	}
	return nil
}

func fetchNextHotshotBlock(
	ctx context.Context,
	espressoClient EspressoClientInterface,
	nextHotshotBlockNum uint64,
	parseHotShotPayloadFn func(tx espressoTypes.Bytes, l1Height uint64) error,
	namespace uint64,
	log log.Logger,
) (uint64, error) {
	latestBlockHeight, err := espressoClient.FetchLatestBlockHeight(ctx)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrFailedToFetchTransactions, err)
	}

	fromBlock := nextHotshotBlockNum
	untilBlock := latestBlockHeight

	if fromBlock > untilBlock {
		return fromBlock, nil
	}

	if untilBlock-fromBlock > HOTSHOT_RANGE_LIMIT {
		untilBlock = nextHotshotBlockNum + HOTSHOT_RANGE_LIMIT
	}

	if fromBlock == untilBlock {
		return fromBlock, nil
	}

	// FetchNamespaceTransactionsInRange is exclusive of the last element.
	namespaceTransactionRangeData, err := espressoClient.FetchNamespaceTransactionsInRange(ctx, fromBlock, untilBlock, namespace)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrFailedToFetchTransactions, err)
	}
	if len(namespaceTransactionRangeData) == 0 {
		// Empty blocks are valid; no transactions is not an error.
		return untilBlock, nil
	}

	// we are subtracting 1 here because FetchNamespaceTransactionsInRange is exclusive of the last element
	header, err := espressoClient.FetchHeaderByHeight(ctx, untilBlock-1)
	l1Height := uint64(0)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrFailedToFetchTransactions, err)
	}

	finalized := header.Header.GetL1Finalized()
	if finalized != nil {
		l1Height = finalized.Number
	}

	for _, namespaceTransactionData := range namespaceTransactionRangeData {
		for _, tx := range namespaceTransactionData.Transactions {
			txPayloadBytes := tx.Payload
			if err := parseHotShotPayloadFn(txPayloadBytes, l1Height); err != nil {
				log.Warn("failed to verify espresso transaction", "err", err)
			}
		}
	}

	return untilBlock, nil
}

func (s *EspressoStreamer) Start(ctxIn context.Context) error {
	ctx, cancel := context.WithCancel(ctxIn)
	s.cancel = cancel

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		debouncer := logDebouncer{
			duration: 3 * time.Minute,
			interval: time.Minute,
		}
		for {
			select {
			case <-ctx.Done():
				s.log.Info("streamer shutting down")
				return
			default:
			}

			prevHotshotBlockNum := s.nextHotshotBlockNum
			err := s.QueueMessagesFromHotshot(ctx, s.parseEspressoTransaction)

			// Integer division detects 1000-block boundary crossings so ranges
			// that skip a multiple of 1000 still trigger the Info log.
			if s.nextHotshotBlockNum/1000 > prevHotshotBlockNum/1000 {
				s.log.Info("Now processing hotshot block", "block number", s.nextHotshotBlockNum)
			} else {
				s.log.Debug("Now processing hotshot block", "block number", s.nextHotshotBlockNum)
			}

			if err != nil {
				if !errors.Is(err, ErrFailedToFetchTransactions) {
					s.log.Error("error while queueing messages from hotshot", "err", err)
				} else if shouldLog, logError := debouncer.debounce(); shouldLog == ShouldLog {
					if logError == ShouldLogAsError {
						s.log.Error("error while queueing messages from hotshot", "err", err)
					} else {
						s.log.Warn("error while queueing messages from hotshot", "err", err)
					}
				}
				time.Sleep(s.retryTime)
			} else {
				debouncer.reset()
			}
		}
	}()

	return nil
}

func (s *EspressoStreamer) StopAndWait() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}
