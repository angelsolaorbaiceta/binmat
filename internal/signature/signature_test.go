package signature

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSignature(t *testing.T) {
	patterns := map[string]string{
		"a": "{ 01 02 ?? 04 }",
		"b": "some string",
	}

	t.Run("Create signature", func(t *testing.T) {
		sig, err := Make("name", "description", patterns, "a AND b")

		assert.Nil(t, err)
		assert.Equal(t, "name", sig.Name)
		assert.Equal(t, "description", sig.Description)
		assert.Equal(t, patterns, sig.Patterns)
		assert.Equal(t, "a AND b", sig.Condition)

		wantPatterns := map[string]*SignaturePattern{
			"a": MakePatternWithMask([]byte{0x01, 0x02, 0x00, 0x04}, []byte{0xff, 0xff, 0x00, 0xff}),
			"b": MakePattern([]byte("some string")),
		}
		assert.Equal(t, wantPatterns, sig.patterns)

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
		assert.Equal(t, ErrSigEmptyName, err.(ErrSignature).reason)
	})

	t.Run("Can't create signature with nil patterns map", func(t *testing.T) {
		_, err := Make("name", "description", nil, "a AND b")

		assert.NotNil(t, err)
		assert.Equal(t, ErrSigEmptyPatterns, err.(ErrSignature).reason)
	})

	t.Run("Can't create signature with empty patterns map", func(t *testing.T) {
		_, err := Make("name", "description", map[string]string{}, "a AND b")

		assert.NotNil(t, err)
		assert.Equal(t, ErrSigEmptyPatterns, err.(ErrSignature).reason)
	})

	t.Run("Can't create signature with a non-parsable pattern", func(t *testing.T) {
		_, err := Make("name", "description", map[string]string{"a": "{ 01 02 b 78 }"}, "a")

		assert.NotNil(t, err)
		assert.Equal(t, ErrSigWrongPattern, err.(ErrSignature).reason)
	})

	t.Run("Can't create signature with empty condition", func(t *testing.T) {
		_, err := Make("name", "description", patterns, "")

		assert.NotNil(t, err)
		assert.Equal(t, ErrSigWrongCondition, err.(ErrSignature).reason)
	})

	t.Run("Can't create signature with a non-parsable condition", func(t *testing.T) {
		_, err := Make("name", "description", patterns, "a AND OR b")

		assert.NotNil(t, err)
		assert.Equal(t, ErrSigWrongCondition, err.(ErrSignature).reason)
	})

	t.Run("Can't create signature with a condition that contains variables not in the patterns", func(t *testing.T) {
		_, err := Make("name", "description", patterns, "a AND (b OR c)")

		assert.NotNil(t, err)
		assert.Equal(t, ErrSigMissingPattern, err.(ErrSignature).reason)
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
	patterns := map[string]string{
		// Pattern a is found at offsets 4 and 13
		"a": "{ 01 02 03 }",
		// Pattern b is found at offset 6
		"b": "{ 03 02 01 }",
		"c": "{ 44 55 66 }",
	}

	t.Run("no match", func(t *testing.T) {
		noMatchSig, _ := Make("test", "test signature", patterns, "a AND (b AND c)")
		matches := noMatchSig.CheckMatch(fileBytes, "path/to/bin")

		assert.False(t, matches.IsMatch)
	})

	t.Run("match", func(t *testing.T) {
		matchSig, _ := Make("test", "test signature", patterns, "a AND (b AND NOT c)")
		matches := matchSig.CheckMatch(fileBytes, "path/to/bin")

		assert.True(t, matches.IsMatch)
		assert.Equal(t, "path/to/bin", matches.FilePath)
		assert.Equal(t, "test", matches.SignatureName)
	})

	t.Run("matches offsets", func(t *testing.T) {
		matchSig, _ := Make("test", "test signature", patterns, "a AND (b AND NOT c)")
		matches := matchSig.CheckMatch(fileBytes, "path/to/bin")

		aOff := matches.OffsetsByPattern["a"]
		assert.Equal(t, PatternMatchOffsets{4, 13}, aOff)

		bOff := matches.OffsetsByPattern["b"]
		assert.Equal(t, PatternMatchOffsets{6}, bOff)

		cOff := matches.OffsetsByPattern["c"]
		assert.Nil(t, cOff)
	})
}

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

		assert.Equal(t, PatternMatchOffsets{0}, matches)
	})

	t.Run("One match starting at offset 2", func(t *testing.T) {
		var (
			data    = []byte{0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, PatternMatchOffsets{2}, matches)
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

		assert.Equal(t, PatternMatchOffsets{0, 5}, matches)
	})

	t.Run("Match ending at the last byte", func(t *testing.T) {
		var (
			data    = []byte{0x00, 0x00, 0x01, 0x02, 0x03}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, PatternMatchOffsets{2}, matches)
	})

	t.Run("Pattern is the whole data", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0x02, 0x03}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, PatternMatchOffsets{0}, matches)
	})

	t.Run("Single byte pattern", func(t *testing.T) {
		var (
			single  = MakePattern([]byte{0x03})
			data    = []byte{0x03, 0x01, 0x03}
			matches = single.checkMatch(data)
		)

		assert.Equal(t, PatternMatchOffsets{0, 2}, matches)
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

		assert.Equal(t, PatternMatchOffsets{0}, matches)
	})

	t.Run("One match starting at offset 2", func(t *testing.T) {
		var (
			data    = []byte{0x00, 0x00, 0x01, 0xab, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		assert.Equal(t, PatternMatchOffsets{2}, matches)
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

		assert.Equal(t, PatternMatchOffsets{0, 5}, matches)
	})
}
