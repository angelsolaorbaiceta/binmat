package signature

const (
	matchByte = 0xff
	anyByte   = 0x00
)

// matchOffsets is a slice of offsets where a pattern matches.
type matchOffsets []int

// len returns the number of offsets.
func (m matchOffsets) len() int {
	return len(m)
}

// isMatch returns true if there is at least one match.
func (m matchOffsets) isMatch() bool {
	return m.len() > 0
}

// A signaturePattern is a single stream of bytes that files are matched against.
type signaturePattern struct {
	pattern []byte
	mask    []byte
	// maskedPattern is the pattern with the mask applied.
	maskedPattern []byte
}

// length returns the length of the pattern and mask.
func (s *signaturePattern) length() int {
	return len(s.pattern)
}

func makePattern(pattern []byte) *signaturePattern {
	mask := make([]byte, len(pattern))
	for i := range mask {
		mask[i] = matchByte
	}

	return makePatternWithMask(pattern, mask)
}

func makePatternWithMask(pattern, mask []byte) *signaturePattern {
	if len(pattern) != len(mask) {
		panic("pattern and mask length mismatch")
	}

	maskedPattern := make([]byte, len(pattern))
	for i := range pattern {
		maskedPattern[i] = pattern[i] & mask[i]
	}

	return &signaturePattern{
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
func (s *signaturePattern) checkMatch(data []byte) matchOffsets {
	var (
		offsets     []int
		fileByte    byte
		patternByte byte

		patternFirstByte = s.pattern[0]
		maskFirstByte    = s.mask[0]
	)

	for i := 0; i < len(data)-s.length(); i++ {
		fileByte = data[i] & maskFirstByte

		if fileByte != patternFirstByte {
			continue
		}

		// The byte at i matches the first byte of the pattern.
		// Check if the rest of the pattern matches.
		for j := 1; j < s.length(); j++ {
			fileByte = data[i+j] & s.mask[j]
			patternByte = s.maskedPattern[j]

			if patternByte != fileByte {
				break
			}

			if j == s.length()-1 {
				offsets = append(offsets, i)
			}
		}
	}

	return offsets
}
