package signature

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignature(t *testing.T) {
	// Simulates the bytes in a binary file. The signatures will be run against
	// these bytes looking for matches.
	fileBytes := []byte{
		// Offset = 0
		0x00, 0x00, 0x00, 0x00,
		// Offset = 4
		0x01, 0x02, 0x03, 0x02, 0x01,
		// Offset = 9
		0x00, 0x00, 0x00, 0x00,
		// Offset = 13
		0x01, 0x02, 0x03, 0x04, 0x05,
		0x00, 0x00, 0x00, 0x00,
	}

	// Pattern a and b are present, but c is not
	patterns := map[string]*signaturePattern{
		// Pattern a is found at offsets 4 and 13
		"a": makePattern([]byte{0x01, 0x02, 0x03}),
		// Pattern b is found at offset 6
		"b": makePattern([]byte{0x03, 0x02, 0x01}),
		"c": makePattern([]byte{0x44, 0x55, 0x66}),
	}

	t.Run("no match", func(t *testing.T) {
		noMatchSig, _ := Make("test", "test signature", patterns, "a AND (b AND c)")
		matches := noMatchSig.CheckMatch(fileBytes)

		assert.False(t, matches.IsMatch)
	})

	t.Run("match", func(t *testing.T) {
		matchSig, _ := Make("test", "test signature", patterns, "a AND (b AND NOT c)")
		matches := matchSig.CheckMatch(fileBytes)

		assert.True(t, matches.IsMatch)
	})

	t.Run("matches offsets", func(t *testing.T) {
		matchSig, _ := Make("test", "test signature", patterns, "a AND (b AND NOT c)")
		matches := matchSig.CheckMatch(fileBytes)

		aOff := matches.Offsets["a"]
		assert.Equal(t, matchOffsets{4, 13}, aOff)

		bOff := matches.Offsets["b"]
		assert.Equal(t, matchOffsets{6}, bOff)

		cOff := matches.Offsets["c"]
		assert.Nil(t, cOff)
	})
}
