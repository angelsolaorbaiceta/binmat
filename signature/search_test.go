package signature

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchMatches(t *testing.T) {
	var (
		dir = t.TempDir()
		one = filepath.Join(dir, "one.bin")
		two = filepath.Join(dir, "two.bin")
	)
	os.WriteFile(one, []byte{0x00, 0x01, 0x02, 0x03}, 0o644)
	os.WriteFile(two, []byte{0xaa, 0xbb, 0xcc}, 0o644)

	// sigA matches "one" only, sigB matches "two" only.
	sigA, _ := Make("sig_a", "", map[string]*SignaturePattern{"p": MakePattern([]byte{0x01, 0x02})}, "p")
	sigB, _ := Make("sig_b", "", map[string]*SignaturePattern{"p": MakePattern([]byte{0xbb, 0xcc})}, "p")
	sigs := Signatures{sigA, sigB}

	byFile := func(matches []SigMatch) map[string]SigMatch {
		m := make(map[string]SigMatch)
		for _, match := range matches {
			m[match.FilePath] = match
		}
		return m
	}

	t.Run("directory: one entry per signature, one result per file", func(t *testing.T) {
		matches, errs := sigs.SearchMatches(dir)

		assert.Empty(t, errs)
		assert.Len(t, matches, 2)

		a := byFile(matches["sig_a"])
		assert.Len(t, a, 2)
		assert.True(t, a[one].IsMatch)
		assert.False(t, a[two].IsMatch)

		b := byFile(matches["sig_b"])
		assert.Len(t, b, 2)
		assert.False(t, b[one].IsMatch)
		assert.True(t, b[two].IsMatch)
	})

	t.Run("single file", func(t *testing.T) {
		matches, errs := sigs.SearchMatches(one)

		assert.Empty(t, errs)
		assert.Len(t, matches, 2)
		assert.Len(t, matches["sig_a"], 1)
		assert.Equal(t, "sig_a", matches["sig_a"][0].SignatureName)
		assert.True(t, matches["sig_a"][0].IsMatch)
		assert.False(t, matches["sig_b"][0].IsMatch)
	})

	t.Run("missing root", func(t *testing.T) {
		matches, errs := sigs.SearchMatches(filepath.Join(dir, "nope"))

		assert.Nil(t, matches)
		assert.Len(t, errs, 1)
	})
}
