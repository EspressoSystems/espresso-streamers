package nitro

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	espressoTypes "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/spf13/pflag"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

const HOTSHOT_RANGE_LIMIT = 100

var (
	ErrFailedToFetchTransactions  = errors.New("failed to fetch transactions")
	ErrPayloadHadNoMessages       = errors.New("ParseHotShotPayload found no messages, the transaction may be empty")
	ErrUserDataHashNot32Bytes     = errors.New("user data hash is not 32 bytes")
	ErrRetryParsingHotShotPayload = errors.New("failed to parse hotshot payload, but will retry")
)

type EspressoStreamerInterface interface {
	Start(ctx context.Context) error
	Next() *MessageWithMetadataAndPos
	// Peek returns the next message in the streamer's buffer. If the message is not
	// in the buffer, it will return nil.
	Peek() *MessageWithMetadataAndPos
	// Advance moves the current message position to the next message.
	Advance()
	// Reset sets the current message position and the next hotshot block number.
	Reset(currentMessagePos uint64, currentHostshotBlock uint64)
	// RecordTimeDurationBetweenHotshotAndCurrentBlock records the time duration between
	// the next hotshot block and the current block.
	RecordTimeDurationBetweenHotshotAndCurrentBlock(nextHotshotBlock uint64, blockProductionTime time.Time)
	GetCurrentEarliestHotShotBlockNumber(pos uint64) uint64

	StopAndWait()
}

type EspressoStreamerConfig struct {
	HotShotBlock        uint64        `koanf:"hotshot-block"`
	TxnsPollingInterval time.Duration `koanf:"txns-polling-interval"`
}

var DefaultEspressoStreamerConfig = EspressoStreamerConfig{
	HotShotBlock: 1,
	// Hotshot currently produces blocks at average of 2 seconds
	// We set it to 1 second to get updates more often than blocks are produced
	TxnsPollingInterval: time.Second,
}

func EspressoStreamerConfigAddOptions(prefix string, f *pflag.FlagSet) {
	f.Uint64(prefix+".hotshot-block", DefaultEspressoStreamerConfig.HotShotBlock, "specifies the hotshot block number to start the espresso streamer on")
	f.Duration(prefix+".txns-polling-interval", DefaultEspressoStreamerConfig.TxnsPollingInterval, "interval between polling for transactions to be included in the block")
}

type EspressoStreamer struct {
	espressoClient            espressoClient.EspressoClient
	nextHotshotBlockNum       uint64
	currentMessagePos         uint64
	namespace                 uint64
	messageWithMetadataAndPos map[uint64]*MessageWithMetadataAndPos
	highestPos                uint64

	messageLock sync.RWMutex
	retryTime   time.Duration

	validBatcherAddresses []common.Address

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

var _ EspressoStreamerInterface = (*EspressoStreamer)(nil)

func NewEspressoStreamer(
	namespace uint64,
	nextHotshotBlockNum uint64,
	espressoClient espressoClient.EspressoClient,
	validBatcherAddresses []common.Address,
	retryTime time.Duration,
) *EspressoStreamer {

	return &EspressoStreamer{
		espressoClient:        espressoClient,
		nextHotshotBlockNum:   nextHotshotBlockNum,
		namespace:             namespace,
		validBatcherAddresses: append([]common.Address(nil), validBatcherAddresses...),
		retryTime:             retryTime,
		currentMessagePos:     1,
	}
}

// GetMessageCount
// This function counts the consecutive positions from `currentMessagePos`.
// Parameters:
//
//	None
//
// Return value:
//
//	a uint64 representing the estimated message count.
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

func (s *EspressoStreamer) Next() *MessageWithMetadataAndPos {
	result := s.Peek()
	if result == nil {
		return nil
	}

	// Advance the current message position, so that the next call to
	// `Peek` or `Next` will return the next message
	s.Advance()
	return result
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

// Call this function to advance the streamer to the next message
func (s *EspressoStreamer) Advance() {
	s.messageLock.Lock()
	defer s.messageLock.Unlock()
	delete(s.messageWithMetadataAndPos, s.currentMessagePos)
	s.currentMessagePos += 1
}

// This function keep fetching hotshot blocks and parsing them until the condition is met.
// It is a do-while loop, which means it will always execute at least once.
//
// Expose the *parseHotShotPayloadFn* to the caller for testing purposes
func (s *EspressoStreamer) QueueMessagesFromHotshot(
	ctx context.Context,
	parseHotShotPayloadFn func(tx espressoTypes.Bytes) error,
) error {
	startHotshotBlockNum := s.nextHotshotBlockNum
	toBlock, err := fetchNextHotshotBlock(
		ctx,
		s.espressoClient,
		startHotshotBlockNum,
		parseHotShotPayloadFn,
		s.namespace,
	)
	if err != nil {
		return err
	}

	s.messageLock.Lock()
	defer s.messageLock.Unlock()
	// Case where we call `reset()` we dont want to jump to `toBlock`
	if s.nextHotshotBlockNum == startHotshotBlockNum {
		s.nextHotshotBlockNum = toBlock
	}
	return nil
}

func (s *EspressoStreamer) verifyBatchPosterSignature(signature []byte, userDataHash [32]byte) error {
	publicKey, err := crypto.SigToPub(userDataHash[:], signature)
	if err != nil {
		return fmt.Errorf("failed to convert signature to public key: %w", err)
	}
	addr := crypto.PubkeyToAddress(*publicKey)
	valid := slices.Contains(s.validBatcherAddresses, addr)
	if !valid {
		log.Error("address not valid", "addr", addr)
		return fmt.Errorf("address not valid: %v", addr)
	}
	return nil
}

func (s *EspressoStreamer) GetCurrentEarliestHotShotBlockNumber(pos uint64) uint64 {
	s.messageLock.RLock()
	defer s.messageLock.RUnlock()
	if len(s.messageWithMetadataAndPos) == 0 {
		// This case means that the espresso streamer is empty and the earliest hotshot block number
		// is the next hotshot block number.
		return s.nextHotshotBlockNum
	}
	if msg, exists := s.messageWithMetadataAndPos[pos]; exists {
		return msg.HotshotHeight
	}

	// Case where pos may not be found, but we have other positions
	minHeight := s.nextHotshotBlockNum
	for nextPos := pos; nextPos <= s.highestPos; nextPos++ {
		if msg, ok := s.messageWithMetadataAndPos[nextPos]; ok && msg.HotshotHeight < minHeight {
			minHeight = msg.HotshotHeight
		}
	}
	return minHeight
}

func (s *EspressoStreamer) parseEspressoTransaction(tx espressoTypes.Bytes) error {
	signature, userDataHash, indices, messages, err := ParseHotShotPayload(tx)
	if err != nil {
		log.Warn("failed to parse hotshot payload", "err", err)
		return err
	}
	if len(messages) == 0 {
		return ErrPayloadHadNoMessages
	}
	if len(userDataHash) != 32 {
		log.Warn("user data hash is not 32 bytes")
		return ErrUserDataHashNot32Bytes
	}

	userDataHashArr := [32]byte(userDataHash)

	err = s.verifyBatchPosterSignature(signature, userDataHashArr)
	if err != nil {
		log.Warn("failed to verify batch poster signature", "err", err)
		return err
	}

	s.messageLock.Lock()
	defer s.messageLock.Unlock()
	for i, message := range messages {
		var messageWithMetadata MessageWithMetadata
		err = rlp.DecodeBytes(message, &messageWithMetadata)
		if err != nil {
			log.Warn("failed to decode message", "err", err)
			// Instead of returnning an error, we should just skip this message
			continue
		}

		if indices[i] < s.currentMessagePos {
			log.Warn("message index is less than current message pos, skipping", "msgPos", indices[i], "currentMessagePos", s.currentMessagePos)
			continue
		}

		msg := &MessageWithMetadataAndPos{
			MessageWithMeta: messageWithMetadata,
			Pos:             indices[i],
			HotshotHeight:   s.nextHotshotBlockNum,
		}

		s.messageWithMetadataAndPos[msg.Pos] = msg

		if msg.Pos > s.highestPos {
			s.highestPos = msg.Pos
		}

		// Check if we have a higher position in an earlier block
		currHeight := msg.HotshotHeight
		for nextPos := msg.Pos + 1; nextPos <= s.highestPos; nextPos++ {
			if higherPos, ok := s.messageWithMetadataAndPos[nextPos]; ok && higherPos.HotshotHeight < currHeight {
				s.messageWithMetadataAndPos[msg.Pos].HotshotHeight = higherPos.HotshotHeight
				currHeight = higherPos.HotshotHeight
			}
		}

		log.Info("Added message to queue", "message", indices[i])
	}
	return nil
}

func (s *EspressoStreamer) RecordTimeDurationBetweenHotshotAndCurrentBlock(nextHotshotBlock uint64, blockProductionTime time.Time) {
	_ = nextHotshotBlock
	_ = blockProductionTime
}

func fetchNextHotshotBlock(
	ctx context.Context,
	espressoClient espressoClient.EspressoClient,
	nextHotshotBlockNum uint64,
	parseHotShotPayloadFn func(tx espressoTypes.Bytes) error,
	namespace uint64,
) (uint64, error) {

	// get the current hotshot block
	latestBlockHeight, err := espressoClient.FetchLatestBlockHeight(ctx)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrFailedToFetchTransactions, err)
	}

	fromBlock := nextHotshotBlockNum
	toBlock := latestBlockHeight

	if latestBlockHeight-nextHotshotBlockNum > HOTSHOT_RANGE_LIMIT {
		toBlock = nextHotshotBlockNum + HOTSHOT_RANGE_LIMIT
	}

	// this means we have no blocks to process and we are all caught up
	if fromBlock == toBlock {
		return toBlock, nil
	}

	// here we are fetching transactions in range [fromBlock, toBlock) exclusive
	//  by default FetchNamespaceTransactionsInRange is exclusive of the last element
	namespaceTransactionRangeData, err := espressoClient.FetchNamespaceTransactionsInRange(ctx, fromBlock, toBlock, namespace)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrFailedToFetchTransactions, err)
	}
	if len(namespaceTransactionRangeData) == 0 {
		// no transactions found in this range is a valid state (e.g., empty blocks), not an error
		return toBlock, nil
	}

	for _, namespaceTransactionData := range namespaceTransactionRangeData {
		for _, tx := range namespaceTransactionData.Transactions {
			txPayloadBytes := tx.Payload
			err := parseHotShotPayloadFn(txPayloadBytes)
			if err != nil && !strings.Contains(err.Error(), ErrRetryParsingHotShotPayload.Error()) {
				log.Warn("failed to verify espresso transaction", "err", err)
				continue
			}
			if err != nil {
				return 0, err
			}
		}
	}

	return toBlock, nil
}

func (s *EspressoStreamer) Start(ctxIn context.Context) error {
	ctx, cancel := context.WithCancel(ctxIn)
	s.cancel = cancel

	const ephemeralDuration = 3 * time.Minute
	const ephemeralLogInterval = 1 * time.Minute

	var (
		ephemeralFirstSeen time.Time
		ephemeralLastLog   time.Time
	)

	s.wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			prevHotshotBlockNum := s.nextHotshotBlockNum

			err := s.QueueMessagesFromHotshot(ctx, s.parseEspressoTransaction)

			// Use integer division to detect 1000-block boundary crossings, so
			// ranges that skip over a multiple of 1000 still trigger the Info log.
			if s.nextHotshotBlockNum/1000 > prevHotshotBlockNum/1000 {
				log.Info("Now processing hotshot block", "block number", s.nextHotshotBlockNum)
			} else {
				log.Debug("Now processing hotshot block", "block number", s.nextHotshotBlockNum)
			}
			if err != nil {
				now := time.Now()
				isEphemeral := errors.Is(err, ErrFailedToFetchTransactions)

				if isEphemeral {
					if ephemeralFirstSeen.IsZero() {
						ephemeralFirstSeen = now
					}
					if time.Since(ephemeralFirstSeen) < ephemeralDuration {
						// Within grace period: downgrade to Warn, rate-limited
						if ephemeralLastLog.IsZero() || time.Since(ephemeralLastLog) >= ephemeralLogInterval {
							log.Warn("error while queueing messages from hotshot", "err", err)
							ephemeralLastLog = now
						}
					} else {
						// Past grace period: escalate to Error
						log.Error("error while queueing messages from hotshot", "err", err)
					}
				} else {
					log.Error("error while queueing messages from hotshot", "err", err)
				}

				select {
				case <-time.After(s.retryTime):
				case <-ctx.Done():
					return
				}
			} else {
				ephemeralFirstSeen = time.Time{}
				ephemeralLastLog = time.Time{}
			}
		}
	})

	return nil
}

func (s *EspressoStreamer) StopAndWait() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}
