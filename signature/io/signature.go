package io

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/angelsolaorbaiceta/binmat/signature"
	"gopkg.in/yaml.v3"
)

var bytePatternRe = regexp.MustCompile(`^\s*\{[0-9a-fA-F ?]*\}\s*$`)

// A Signature is the serialization read/write entity for a domain signature.
type Signature struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Patterns    map[string]string `yaml:"patterns"`
	Condition   string            `yaml:"condition"`
}

// ReadFromYaml attempts to decode a Signature from a yaml file.
func ReadFromYaml(r io.Reader) (Signature, error) {
	var (
		decoder   = yaml.NewDecoder(r)
		signature Signature
		err       = decoder.Decode(&signature)
	)

	return signature, err
}

// ToDomain maps the signature to a domain instance of the signature.
// The returned error can be:
//   - ErrSignature: if the error happens in the creation of the signature
//   - error: if there's an error parsing a pattern
func (s Signature) ToDomain() (signature.Signature, error) {
	var (
		patterns = make(map[string]*signature.SignaturePattern)
		err      error
	)

	for name, pattern := range s.Patterns {
		patterns[name], err = patternToDomain(pattern)
		if err != nil {
			return signature.Signature{}, err
		}
	}

	return signature.Make(s.Name, s.Description, patterns, s.Condition)
}

// patternToDomain parses a given pattern into a domain SignaturePattern.
// Patterns can be binary sequences or strings.
// Returns an error if the pattern can't be parsed.
func patternToDomain(pattern string) (*signature.SignaturePattern, error) {
	if bytePatternRe.MatchString(pattern) {
		var (
			stripped    = strings.Trim(pattern, "{ }")
			fields      = strings.Fields(stripped)
			bytePattern = make([]byte, len(fields))
			byteMask    = make([]byte, len(fields))
		)

		for i, field := range fields {
			if len(field) != 2 {
				return nil, fmt.Errorf("byte should have a length of 2 chars, got '%s'", field)
			}

			if field == "??" {
				bytePattern[i] = 0x00
				byteMask[i] = 0x00
			} else {
				// At this point, field is known to be a two characters string consisting
				// of numbers and the letters A to F, thus the ParseUInt using hexadecimal
				// base can't fail. The error is ignored.
				value, _ := strconv.ParseUint(field, 16, 8)
				bytePattern[i] = byte(value)
				byteMask[i] = 0xff
			}
		}

		return signature.MakePatternWithMask(bytePattern, byteMask), nil
	}

	// The sequence appears to be a string. Convert to its ascii bytes.
	return signature.MakePattern([]byte(pattern)), nil
}
