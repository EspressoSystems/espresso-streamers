package nitro

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogDebouncer(t *testing.T) {
	t.Run("first observe always logs as warn", func(t *testing.T) {
		l := logDebouncer{duration: time.Minute, interval: time.Minute}
		shouldLog, asError := l.debounce()
		assert.Equal(t, ShouldLog, shouldLog)
		assert.Equal(t, ShouldNotLogAsError, asError)
	})

	t.Run("second observe within interval is suppressed", func(t *testing.T) {
		l := logDebouncer{duration: time.Minute, interval: time.Minute}
		l.debounce()
		shouldLog, _ := l.debounce()
		assert.Equal(t, ShouldNotLog, shouldLog)
	})

	t.Run("escalates to error after grace period", func(t *testing.T) {
		l := logDebouncer{duration: -1, interval: time.Minute}
		shouldLog, asError := l.debounce()
		assert.Equal(t, ShouldLog, shouldLog)
		assert.Equal(t, ShouldLogAsError, asError)
	})

	t.Run("reset clears state so next observe logs as warn", func(t *testing.T) {
		l := logDebouncer{duration: time.Minute, interval: time.Minute}
		l.debounce()
		l.reset()
		shouldLog, asError := l.debounce()
		assert.Equal(t, ShouldLog, shouldLog)
		assert.Equal(t, ShouldNotLogAsError, asError)
	})

	t.Run("reset preserves duration and interval", func(t *testing.T) {
		l := logDebouncer{duration: 5 * time.Minute, interval: 30 * time.Second}
		l.debounce()
		l.reset()
		assert.Equal(t, 5*time.Minute, l.duration)
		assert.Equal(t, 30*time.Second, l.interval)
	})
}
