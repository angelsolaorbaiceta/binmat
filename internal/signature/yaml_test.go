package signature

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignatureYAML(t *testing.T) {
	fileBytes, err := os.ReadFile("../../examples/signatures/__io_test.yaml")
	if err != nil {
		panic("Can't read file" + err.Error())
	}

	t.Run("decodes the textual definition", func(t *testing.T) {
		sig, err := ReadFromYaml(bytes.NewReader(fileBytes))

		assert.Nil(t, err)
		assert.Equal(t, "A test signature", sig.Name)
		assert.Equal(t, "This signature is used in tests", sig.Description)
		assert.Equal(t, map[string]string{
			"a": "{ 74 fc ff ff c6 05 19 45 }",
			"b": " { 22 33 ?? 55 aa bb } ",
			"c": "very wow, much cool",
		}, sig.Patterns)
		assert.Equal(t, "a AND (b AND c)", sig.Condition)
	})

	t.Run("decoding compiles the patterns and condition", func(t *testing.T) {
		sig, _ := ReadFromYaml(bytes.NewReader(fileBytes))

		wantPatterns := map[string]*SignaturePattern{
			"a": MakePattern([]byte{0x74, 0xfc, 0xff, 0xff, 0xc6, 0x05, 0x19, 0x45}),
			"b": MakePatternWithMask(
				[]byte{0x22, 0x33, 0x00, 0x55, 0xaa, 0xbb},
				[]byte{0xff, 0xff, 0x00, 0xff, 0xff, 0xff},
			),
			"c": MakePattern([]byte("very wow, much cool")),
		}
		assert.Equal(t, wantPatterns, sig.patterns)

		result, err := sig.conditionFn(map[string]bool{"a": true, "b": true, "c": false})
		assert.Nil(t, err)
		assert.False(t, result)
	})

	t.Run("decoding fails on an invalid definition", func(t *testing.T) {
		doc := "name: bad\npatterns:\n  a: '{ 01 02 b 78 }'\ncondition: a\n"
		_, err := ReadFromYaml(strings.NewReader(doc))

		var sigErr ErrSignature
		assert.ErrorAs(t, err, &sigErr)
		assert.Equal(t, ErrSigWrongPattern, sigErr.reason)
	})

	t.Run("round trips through WriteYAML", func(t *testing.T) {
		original, _ := ReadFromYaml(bytes.NewReader(fileBytes))

		var buf bytes.Buffer
		assert.Nil(t, original.WriteYAML(&buf))

		decoded, err := ReadFromYaml(&buf)
		assert.Nil(t, err)
		// conditionFn is a func, which is never deep-equal; compare the rest.
		assert.Equal(t, original.Name, decoded.Name)
		assert.Equal(t, original.Description, decoded.Description)
		assert.Equal(t, original.Patterns, decoded.Patterns)
		assert.Equal(t, original.Condition, decoded.Condition)
		assert.Equal(t, original.patterns, decoded.patterns)
		assert.NotNil(t, decoded.conditionFn)
	})
}

func TestParsePattern(t *testing.T) {
	t.Run("byte sequence", func(t *testing.T) {
		p, err := ParsePattern("{ 01 ab ?? FF }")

		assert.Nil(t, err)
		assert.Equal(t, MakePatternWithMask([]byte{0x01, 0xab, 0x00, 0xff}, []byte{0xff, 0xff, 0x00, 0xff}), p)
	})

	t.Run("string", func(t *testing.T) {
		p, err := ParsePattern("hello")

		assert.Nil(t, err)
		assert.Equal(t, MakePattern([]byte("hello")), p)
	})

	t.Run("rejects malformed bytes", func(t *testing.T) {
		_, err := ParsePattern("{ 01 2 }")
		assert.NotNil(t, err)
	})

	t.Run("rejects empty patterns", func(t *testing.T) {
		_, err := ParsePattern("")
		assert.NotNil(t, err)

		_, err = ParsePattern("{ }")
		assert.NotNil(t, err)
	})
}
