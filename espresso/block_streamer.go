package espresso

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"

	esptypes "github.com/EspressoSystems/espresso-network/sdks/go/types"
	espcommon "github.com/EspressoSystems/espresso-network/sdks/go/types/common"
	espstreamers "github.com/EspressoSystems/espresso-streamers"
)

type PayloadQueryDataStreamer interface {
	espstreamers.Streamer[espcommon.PayloadQueryData]
}

type payloadQueryDataRawHTTPGetStreamer struct {
	client *http.Client
	url    url.URL
	start  uint64
}

func (s *payloadQueryDataRawHTTPGetStreamer) Next(ctx context.Context) (espcommon.PayloadQueryData, error) {
	value, err := performJSONGetRequest[espcommon.PayloadQueryData](
		ctx,
		s.url.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/availability/payload/%d", s.start)}),
	)
	if err != nil {
		return espcommon.PayloadQueryData{}, err
	}

	s.start++
	return value, nil
}

func NewPayloadQueryDataStreamer(u url.URL, start uint64) PayloadQueryDataStreamer {
	return &payloadQueryDataRawHTTPGetStreamer{
		url:   u,
		start: start,
	}
}

func performJSONGetRequest[R any](ctx context.Context, u *url.URL) (result R, err error) {
	request := http.Request{
		Method: http.MethodGet,
		URL:    u,
		Header: http.Header{
			"Accept": []string{"application/json"},
		},
	}

	res, err := http.DefaultClient.Do(&request)
	if err != nil {
		return result, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		// TODO: return error
		return result, errors.New("failed to fetch entry")
	}

	mediaType, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		return result, err
	}

	if mediaType != "application/json" {
		// TODO: return error
		return result, errors.New("unsupported content-type")
	}

	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&result)
	return result, err
}

const (
	len32 = 4
)

type namespaceRangeStreamer struct {
	nsTable esptypes.NsTable
	start   uint32
	offset  uint32
}

func (s *namespaceRangeStreamer) Next(ctx context.Context) (NamespaceRange, error) {
	l := uint32(len(s.nsTable.Bytes))
	if l < len32+len32 || l-len32-len32 < s.offset {
		return NamespaceRange{}, espstreamers.ErrEndOfStream
	}

	if s.offset == 0 {
		// we're at the first byte, but the payload actually starts at the 4th
		// byte.
		s.offset = len32
		return s.Next(ctx)
	}

	// we read two uint32s stored in big endian.  The first is the namespace,
	// the second is the end byte of the payload range.
	namespace := binary.LittleEndian.Uint32(s.nsTable.Bytes[s.offset : s.offset+len32])
	end := binary.LittleEndian.Uint32(s.nsTable.Bytes[s.offset+len32 : s.offset+len32+len32])
	start := s.start
	s.start = end
	s.offset += len32 + len32

	return NamespaceRange{
		Namespace: namespace,
		ByteRange: ByteRange{
			Start: start,
			End:   end,
		},
	}, nil
}

type NamespaceRange struct {
	Namespace uint32
	ByteRange
}

type ByteRange struct {
	Start uint32
	End   uint32
}

type PayloadRangeStreamer struct {
	payload []byte
	start   uint32
	offset  uint32
	entries uint32
}

func (s *PayloadRangeStreamer) Next(ctx context.Context) (ByteRange, error) {
	l := uint32(len(s.payload))
	if l < len32 || l-len32 < s.offset {
		return ByteRange{}, espstreamers.ErrEndOfStream
	}

	if s.offset == 0 {
		// We're at the first byte, we don't know how many entries there are yet.
		// the first uint32 indicates how many entries there are, so we read that
		// and then move the offset forward.

		numEntries := binary.LittleEndian.Uint32(s.payload[:len32])
		s.entries = numEntries
		s.offset += len32
		s.start = numEntries*len32 + len32
		return s.Next(ctx)
	}

	if (s.offset / len32) > s.entries+1 {
		// We've exhausted all the entries, so we return end of stream.
		return ByteRange{}, espstreamers.ErrEndOfStream
	}

	end := binary.LittleEndian.Uint32(s.payload[s.offset : s.offset+len32])
	start := s.start
	s.start = end
	s.offset += len32

	return ByteRange{
		Start: start,
		End:   end,
	}, nil
}

func (r ByteRange) slice(payload []byte) []byte {
	if payload == nil {
		return nil
	}

	start, end := r.Start, r.End
	payloadLen := uint32(len(payload))
	if end > payloadLen {
		end = payloadLen
	}

	if start > payloadLen {
		start = payloadLen
	}

	return payload[r.Start:r.End]
}

type transactionFromBlockDataStreamer struct {
	payload                []byte
	payloadRangeStreamer   espstreamers.PeekStreamer[ByteRange]
	namespaceRangeStreamer espstreamers.PeekStreamer[NamespaceRange]
}

func (s *transactionFromBlockDataStreamer) Next(ctx context.Context) (espcommon.Transaction, error) {
	nsRange, err := s.namespaceRangeStreamer.Next(ctx)
	if err != nil {
		return espcommon.Transaction{}, err
	}
	bRange, err := s.payloadRangeStreamer.Peek(ctx)
	if err != nil {
		return espcommon.Transaction{}, err
	}

	for bRange.End > nsRange.End {
		_, _ = s.namespaceRangeStreamer.Next(ctx)
		nsRange, err = s.namespaceRangeStreamer.Peek(ctx)
		if err != nil {
			return espcommon.Transaction{}, err
		}
	}

	return espcommon.Transaction{
		Namespace: uint64(nsRange.Namespace),
		Payload:   espcommon.Bytes(bRange.slice(s.payload)),
	}, nil
}

func TransactionStreamerFromPayloadData(payload *espcommon.BlockPayload) espstreamers.Streamer[espcommon.Transaction] {
	return &transactionFromBlockDataStreamer{
		payload:                payload.RawPayload,
		payloadRangeStreamer:   espstreamers.AddPeek(&PayloadRangeStreamer{payload: payload.RawPayload}),
		namespaceRangeStreamer: espstreamers.AddPeek(&namespaceRangeStreamer{nsTable: payload.NsTable}),
	}
}
