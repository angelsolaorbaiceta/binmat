package signature

const (
	maskMatchByte = 0xff
	maskAnyByte   = 0x00
)

// MatchOffsets is a slice of offsets where a pattern matches.
type MatchOffsets []int

// len returns the number of offsets.
func (m MatchOffsets) len() int {
	return len(m)
}

// isMatch returns true if there is at least one match.
func (m MatchOffsets) isMatch() bool {
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
func (s *SignaturePattern) checkMatch(data []byte) MatchOffsets {
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
