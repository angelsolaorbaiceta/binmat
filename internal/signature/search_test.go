package signature

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchMatches(t *testing.T) {
	var (
		dir   = t.TempDir()
		one   = filepath.Join(dir, "one.bin")
		two   = filepath.Join(dir, "two.bin")
		three = filepath.Join(dir, "three.bin")
	)
	os.WriteFile(one, []byte{0x00, 0x01, 0x02, 0x03}, 0o644)
	os.WriteFile(two, []byte{0xaa, 0xbb, 0xcc}, 0o644)
	os.WriteFile(three, []byte{0xff, 0xff, 0xff}, 0o644)

	// sigA matches "one" only, sigB matches "two" only, nothing matches "three".
	sigA, _ := Make("sig_a", "", map[string]string{"p": "{ 01 02 }"}, "p")
	sigB, _ := Make("sig_b", "", map[string]string{"p": "{ bb cc }"}, "p")
	sigs := Signatures{sigA, sigB}

	t.Run("directory: only matching results are kept, grouped by signature", func(t *testing.T) {
		matches, errs := sigs.SearchMatches(dir)

		assert.Empty(t, errs)
		assert.Len(t, matches, 2)

		assert.Len(t, matches["sig_a"], 1)
		assert.Equal(t, one, matches["sig_a"][0].FilePath)
		assert.Equal(t, "sig_a", matches["sig_a"][0].SignatureName)
		assert.True(t, matches["sig_a"][0].IsMatch)

		assert.Len(t, matches["sig_b"], 1)
		assert.Equal(t, two, matches["sig_b"][0].FilePath)
		assert.Equal(t, "sig_b", matches["sig_b"][0].SignatureName)
		assert.True(t, matches["sig_b"][0].IsMatch)
	})

	t.Run("single file: signatures that don't match have no entry", func(t *testing.T) {
		matches, errs := sigs.SearchMatches(one)

		assert.Empty(t, errs)
		assert.Len(t, matches, 1)
		assert.Len(t, matches["sig_a"], 1)
		assert.Equal(t, one, matches["sig_a"][0].FilePath)
		assert.NotContains(t, matches, "sig_b")
	})

	t.Run("single file with no matches yields an empty map", func(t *testing.T) {
		matches, errs := sigs.SearchMatches(three)

		assert.Empty(t, errs)
		assert.NotNil(t, matches)
		assert.Empty(t, matches)
	})

	t.Run("missing root", func(t *testing.T) {
		matches, errs := sigs.SearchMatches(filepath.Join(dir, "nope"))

		assert.Nil(t, matches)
		assert.Len(t, errs, 1)
	})
}
