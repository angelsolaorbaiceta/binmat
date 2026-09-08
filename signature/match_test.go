package signature

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchPatternWithoutMask(t *testing.T) {
	sig := MakePattern([]byte{0x01, 0x02, 0x03})

	t.Run("No enough bytes to match", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0x02}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, 0, matches.len())
	})

	t.Run("One match starting at offset 0", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0x02, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, MatchOffsets{0}, matches)
	})

	t.Run("One match starting at offset 2", func(t *testing.T) {
		var (
			data    = []byte{0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, MatchOffsets{2}, matches)
	})

	t.Run("No match", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0x02, 0x04, 0x05, 0x06, 0x07}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, 0, matches.len())
	})

	t.Run("Multiple matches", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, MatchOffsets{0, 5}, matches)
	})

	t.Run("Match ending at the last byte", func(t *testing.T) {
		var (
			data    = []byte{0x00, 0x00, 0x01, 0x02, 0x03}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, MatchOffsets{2}, matches)
	})

	t.Run("Pattern is the whole data", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0x02, 0x03}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, MatchOffsets{0}, matches)
	})

	t.Run("Single byte pattern", func(t *testing.T) {
		var (
			single  = MakePattern([]byte{0x03})
			data    = []byte{0x03, 0x01, 0x03}
			matches = single.checkMatch(data)
		)

		assert.Equal(t, MatchOffsets{0, 2}, matches)
	})
}

func TestMatchPatternWithMask(t *testing.T) {
	sig := MakePatternWithMask(
		[]byte{0x01, 0x02, 0x03},
		[]byte{maskMatchByte, maskAnyByte, maskMatchByte},
	)

	t.Run("No enough bytes to match", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0x02}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, 0, matches.len())
	})

	t.Run("One match starting at offset 0", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0xab, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, MatchOffsets{0}, matches)
	})

	t.Run("One match starting at offset 2", func(t *testing.T) {
		var (
			data    = []byte{0x00, 0x00, 0x01, 0xab, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, MatchOffsets{2}, matches)
	})

	t.Run("No match", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0xab, 0x04, 0x05, 0x06, 0x07}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, 0, matches.len())
	})

	t.Run("Multiple matches", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0xab, 0x03, 0x04, 0x05, 0x01, 0xcd, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, MatchOffsets{0, 5}, matches)
	})
}
