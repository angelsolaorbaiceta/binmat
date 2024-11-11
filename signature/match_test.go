package signature

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchPatternWithoutMask(t *testing.T) {
	sig := makePattern([]byte{0x01, 0x02, 0x03})

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

		assert.Equal(t, matchOffsets{0}, matches)
	})

	t.Run("One match starting at offset 2", func(t *testing.T) {
		var (
			data    = []byte{0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, matchOffsets{2}, matches)
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

		assert.Equal(t, matchOffsets{0, 5}, matches)
	})
}

func TestMatchPatternWithMask(t *testing.T) {
	sig := makePatternWithMask(
		[]byte{0x01, 0x02, 0x03},
		[]byte{matchByte, anyByte, matchByte},
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

		assert.Equal(t, matchOffsets{0}, matches)
	})

	t.Run("One match starting at offset 2", func(t *testing.T) {
		var (
			data    = []byte{0x00, 0x00, 0x01, 0xab, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, 2, matches.len())
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

		assert.Equal(t, matchOffsets{0, 5}, matches)
	})
}
