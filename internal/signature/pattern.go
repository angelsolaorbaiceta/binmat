package signature

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	maskMatchByte = 0xff
	maskAnyByte   = 0x00
)

// bytePatternRe recognizes byte sequence patterns: hex byte pairs or "??"
// wildcards, separated by spaces and enclosed in curly brackets.
var bytePatternRe = regexp.MustCompile(`^\s*\{[0-9a-fA-F ?]*\}\s*$`)

// PatternMatchOffsets is a slice of offsets where a pattern matches.
type PatternMatchOffsets []int

// len returns the number of offsets.
func (m PatternMatchOffsets) len() int {
	return len(m)
}

// isMatch returns true if there is at least one match.
func (m PatternMatchOffsets) isMatch() bool {
	return m.len() > 0
}

// A SignaturePattern is a single stream of bytes that files are matched against.
type SignaturePattern struct {
	pattern []byte
	mask    []byte
	// maskedPattern is the pattern with the mask applied.
	maskedPattern []byte
}

// Length returns the Length of the pattern and mask.
func (s *SignaturePattern) Length() int {
	return len(s.pattern)
}

// ParsePattern parses the textual form of a pattern, as written in the
// signature files, into a SignaturePattern.
//
// A pattern is either a byte sequence, like "{ 74 fc ?? 45 }", where "??"
// matches any byte at that position, or any other non-empty string, which is
// matched as its ASCII bytes.
func ParsePattern(pattern string) (*SignaturePattern, error) {
	if !bytePatternRe.MatchString(pattern) {
		if len(pattern) == 0 {
			return nil, errors.New("pattern can't be empty")
		}

		// The sequence appears to be a string. Convert to its ascii bytes.
		return MakePattern([]byte(pattern)), nil
	}

	var (
		stripped    = strings.Trim(pattern, "{ }")
		fields      = strings.Fields(stripped)
		bytePattern = make([]byte, len(fields))
		byteMask    = make([]byte, len(fields))
	)

	if len(fields) == 0 {
		return nil, errors.New("byte sequence can't be empty")
	}

	for i, field := range fields {
		if len(field) != 2 {
			return nil, fmt.Errorf("byte should have a length of 2 chars, got '%s'", field)
		}

		if field == "??" {
			bytePattern[i] = 0x00
			byteMask[i] = maskAnyByte
		} else {
			// At this point, field is known to be a two characters string consisting
			// of numbers and the letters A to F, thus the ParseUInt using hexadecimal
			// base can't fail. The error is ignored.
			value, _ := strconv.ParseUint(field, 16, 8)
			bytePattern[i] = byte(value)
			byteMask[i] = maskMatchByte
		}
	}

	return MakePatternWithMask(bytePattern, byteMask), nil
}

func MakePattern(pattern []byte) *SignaturePattern {
	mask := make([]byte, len(pattern))
	for i := range mask {
		mask[i] = maskMatchByte
	}

	return MakePatternWithMask(pattern, mask)
}

func MakePatternWithMask(pattern, mask []byte) *SignaturePattern {
	if len(pattern) != len(mask) {
		panic("pattern and mask length mismatch")
	}

	maskedPattern := make([]byte, len(pattern))
	for i := range pattern {
		maskedPattern[i] = pattern[i] & mask[i]
	}

	return &SignaturePattern{
		pattern:       pattern,
		mask:          mask,
		maskedPattern: maskedPattern,
	}
}

// checkMatch reads the file from the byte slice and checks if the signature matches.
// It returns all the offsets where the signature matches.
//
// The function expects the full file contents in a byte slice, as binaries themselves
// are usually small enough to fit in memory.
func (s *SignaturePattern) checkMatch(data []byte) PatternMatchOffsets {
	var (
		offsets     []int
		fileByte    byte
		patternByte byte

		patternFirstByte = s.pattern[0]
		maskFirstByte    = s.mask[0]
	)

	// The last candidate offset is the one where the pattern ends exactly at
	// the last byte of the data, hence the <=.
	for i := 0; i <= len(data)-s.Length(); i++ {
		fileByte = data[i] & maskFirstByte

		if fileByte != patternFirstByte {
			continue
		}

		// The byte at i matches the first byte of the pattern.
		// Check if the rest of the pattern matches.
		matched := true
		for j := 1; j < s.Length(); j++ {
			fileByte = data[i+j] & s.mask[j]
			patternByte = s.maskedPattern[j]

			if patternByte != fileByte {
				matched = false
				break
			}
		}

		if matched {
			offsets = append(offsets, i)
		}
	}

	return offsets
}
