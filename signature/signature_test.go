package signature

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSignature(t *testing.T) {
	patterns := map[string]*SignaturePattern{
		"a": nil,
		"b": nil,
	}

	t.Run("Create signature", func(t *testing.T) {
		sig, err := Make("name", "description", patterns, "a AND b")

		assert.Nil(t, err)
		assert.Equal(t, "name", sig.Name)
		assert.Equal(t, "description", sig.Description)
		assert.Equal(t, patterns, sig.Patterns)
		assert.Equal(t, "a AND b", sig.Condition)

		vars := map[string]bool{
			"a": true,
			"b": true,
		}
		result, _ := sig.conditionFn(vars)

		assert.True(t, result)
	})

	t.Run("Can't create signature with empty name", func(t *testing.T) {
		_, err := Make("", "description", patterns, "a AND b")

		assert.NotNil(t, err)
		assert.Equal(t, ErrSigEmptyName, err.reason)
	})

	t.Run("Can't create signature with nil patterns map", func(t *testing.T) {
		_, err := Make("name", "description", nil, "a AND b")

		assert.NotNil(t, err)
		assert.Equal(t, ErrSigEmptyPatterns, err.reason)
	})

	t.Run("Can't create signature with empty patterns map", func(t *testing.T) {
		_, err := Make("name", "description", map[string]*SignaturePattern{}, "a AND b")

		assert.NotNil(t, err)
		assert.Equal(t, ErrSigEmptyPatterns, err.reason)
	})

	t.Run("Can't create signature with empty condition", func(t *testing.T) {
		_, err := Make("name", "description", patterns, "")

		assert.NotNil(t, err)
		assert.Equal(t, ErrSigWrongCondition, err.reason)
	})

	t.Run("Can't create signature with a non-parsable condition", func(t *testing.T) {
		_, err := Make("name", "description", patterns, "a AND OR b")

		assert.NotNil(t, err)
		assert.Equal(t, ErrSigWrongCondition, err.reason)
	})

	t.Run("Can't create signature with a condition that contains variables not in the patterns", func(t *testing.T) {
		_, err := Make("name", "description", patterns, "a AND (b OR c)")

		assert.NotNil(t, err)
		assert.Equal(t, ErrSigMissingPattern, err.reason)
	})
}

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
	patterns := map[string]*SignaturePattern{
		// Pattern a is found at offsets 4 and 13
		"a": MakePattern([]byte{0x01, 0x02, 0x03}),
		// Pattern b is found at offset 6
		"b": MakePattern([]byte{0x03, 0x02, 0x01}),
		"c": MakePattern([]byte{0x44, 0x55, 0x66}),
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
