package signature

const (
	MatchByte = 0xff
	AnyByte   = 0x00
)

// A Signature is a pattern that can be matched in a file.
// A Signature is defined by a name, a description, a pattern, and a mask.
// The pattern is the sequence of bytes that must be matched.
// The mask is applied to the pattern to define which bytes must be matched, and which can be ignored.
type Signature struct {
	Name        string
	Description string
	pattern     []byte
	mask        []byte
	// maskedPattern is the pattern with the mask applied.
	maskedPattern []byte
}

// Make creates a new Signature with the given name, description, and pattern.
// The mask is set to MatchByte for all bytes in the pattern.
func Make(name, description string, pattern []byte) *Signature {
	mask := make([]byte, len(pattern))
	for i := range mask {
		mask[i] = MatchByte
	}

	return MakeWithMask(name, description, pattern, mask)
}

// MakeWithMask creates a new Signature with the given name, description,
// pattern, and mask. The pattern and mask must be the same length.
func MakeWithMask(name, description string, pattern, mask []byte) *Signature {
	if len(pattern) != len(mask) {
		panic("pattern and mask length mismatch")
	}

	maskedPattern := make([]byte, len(pattern))
	for i := range pattern {
		maskedPattern[i] = pattern[i] & mask[i]
	}

	return &Signature{
		Name:          name,
		Description:   description,
		pattern:       pattern,
		mask:          mask,
		maskedPattern: maskedPattern,
	}
}

// length returns the length of the pattern and mask.
func (s *Signature) length() int {
	return len(s.pattern)
}

// CheckMatch reads the file from the byte slice and checks if the signature matches.
// It returns all the offsets where the signature matches.
//
// The function expects the full file contents in a byte slice, as binaries themselves
// are usually small enough to fit in memory.
func (s *Signature) CheckMatch(data []byte) *SigMatches {
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

	return &SigMatches{Offsets: offsets, Signature: s}
}
