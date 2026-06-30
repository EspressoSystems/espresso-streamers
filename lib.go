package espressostreamers

import (
	"context"
	"errors"
)

// ErrEndOfStream is an error that indicates that the streamer has reached
// the end of the stream and there are no more items to read.
var ErrEndOfStream = errors.New("end of stream")

// Streamer is an interface that allows streaming items of type T.
type Streamer[T any] interface {
	// Next returns the next item in the stream.  If there are no
	// more items to produce, and the stream has ended, Next should return
	// [ErrEndOfStream].
	Next(ctx context.Context) (T, error)
}

// Peeker is an interface that allows peeking at the next item in the stream
// without consuming it.
type Peeker[T any] interface {
	// Peek returns the next item in the stream without advancing the collection.
	// If there are no more items to produce, and the collection is at its
	// en, Peek should return [ErrEndOfStream].
	Peek(ctx context.Context) (T, error)
}

type PeekStreamer[T any] interface {
	Peeker[T]
	Streamer[T]
}

func AddPeek[T any](s Streamer[T]) PeekStreamer[T] {
	return &extendPeeker[T]{
		streamer: s,
	}
}

type lastItem[T any] struct {
	Item  T
	Valid bool
}

type extendPeeker[T any] struct {
	streamer Streamer[T]
	last     lastItem[T]
}

func (p *extendPeeker[T]) Next(ctx context.Context) (T, error) {
	item, err := p.Peek(ctx)
	p.last = lastItem[T]{}
	return item, err
}

func (p *extendPeeker[T]) Peek(ctx context.Context) (T, error) {
	if p.last.Valid {
		return p.last.Item, nil
	}

	next, err := p.streamer.Next(ctx)
	if err != nil {
		return next, err
	}

	p.last = lastItem[T]{
		Item:  next,
		Valid: true,
	}

	return next, nil
}

type streamMap[T, U any] struct {
	ts Streamer[T]
	fn func(T) U
}

func (m streamMap[T, U]) Next(ctx context.Context) (result U, err error) {
	next, err := m.ts.Next(ctx)
	if err != nil {
		return result, err
	}
	return m.fn(next), nil
}

func Map[T, U any](ts Streamer[T], fn func(T) U) Streamer[U] {
	return streamMap[T, U]{ts: ts, fn: fn}
}

type streamWhere[T any] struct {
	ts     Streamer[T]
	accept func(T) bool
}

func (w streamWhere[T]) Next(ctx context.Context) (result T, err error) {
	for {
		next, err := w.ts.Next(ctx)
		if err != nil {
			return result, err
		}

		if w.accept(next) {
			return next, nil
		}
	}
}

func Where[T any](ts Streamer[T], accept func(T) bool) Streamer[T] {
	return streamWhere[T]{ts: ts, accept: accept}
}

type EnumeratedEntry[T any] struct {
	Pos   uint64
	Value T
}

type streamEnumerate[T any] struct {
	ts    Streamer[T]
	count uint64
}

func (e *streamEnumerate[T]) Next(ctx context.Context) (result EnumeratedEntry[T], err error) {
	next, err := e.ts.Next(ctx)
	if err != nil {
		return result, err
	}

	defer func() {
		e.count++
	}()

	return EnumeratedEntry[T]{
		Pos:   e.count,
		Value: next,
	}, nil
}

func Enumerate[T any](ts Streamer[T], start uint64) Streamer[EnumeratedEntry[T]] {
	return &streamEnumerate[T]{ts: ts, count: start}
}
