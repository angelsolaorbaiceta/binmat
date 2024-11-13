package io

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/angelsolaorbaiceta/binmat/signature"
	"github.com/stretchr/testify/assert"
)

func TestIOSignature(t *testing.T) {
	fileBytes, err := os.ReadFile("../../examples/signatures/__io_test.yaml")
	if err != nil {
		panic("Can't read file" + err.Error())
	}

	getReader := func() io.Reader {
		return strings.NewReader(string(fileBytes))
	}

	t.Run("Parse signature from YAML", func(t *testing.T) {
		sig, err := ReadFromYaml(getReader())
		want := Signature{
			Name:        "A test signature",
			Description: "This signature is used in tests",
			Patterns: map[string]string{
				"a": "{ 74 fc ff ff c6 05 19 45 }",
				"b": " { 22 33 ?? 55 aa bb } ",
				"c": "very wow, much cool",
			},
			Condition: "a AND (b AND c)",
		}

		assert.Nil(t, err)
		assert.Equal(t, want, sig)
	})

	t.Run("Signature to domain", func(t *testing.T) {
		ioSig, _ := ReadFromYaml(getReader())
		sig, err := ioSig.ToDomain()

		assert.Nil(t, err)
		assert.Equal(t, "A test signature", sig.Name)
		assert.Equal(t, "This signature is used in tests", sig.Description)

		wantPatterns := map[string]*signature.SignaturePattern{
			"a": signature.MakePattern(
				[]byte{0x74, 0xfc, 0xff, 0xff, 0xc6, 0x05, 0x19, 0x45},
			),
			"b": signature.MakePatternWithMask(
				[]byte{0x22, 0x33, 0x00, 0x55, 0xaa, 0xbb},
				[]byte{0xff, 0xff, 0x00, 0xff, 0xff, 0xff},
			),
			"c": signature.MakePattern(
				[]byte{0x76, 0x65, 0x72, 0x79, 0x20, 0x77, 0x6f, 0x77, 0x2c, 0x20, 0x6d,
					0x75, 0x63, 0x68, 0x20, 0x63, 0x6f, 0x6f, 0x6c},
			),
		}
		assert.Equal(t, wantPatterns, sig.Patterns)

		assert.Equal(t, "a AND (b AND c)", sig.Condition)
	})

	t.Run("to domain fails if any field in the string doesn't have two characters", func(t *testing.T) {
		ioSig := Signature{
			Name:        "foo",
			Description: "bar",
			Condition:   "a",
			Patterns:    map[string]string{"a": "{ 01 02 b 78 }"},
		}
		_, err := ioSig.ToDomain()

		assert.NotNil(t, err)
	})
}
