package signature

import "io"

const (
	MatchByte = 0xff
	AnyByte   = 0x00
)

// buffSize is the size of the buffer used to read the file.
// It's a multiple of the OS page size.
// The typical page size is 4096 bytes.
var buffSize int

type Signature struct {
	Name          string
	Description   string
	pattern       []byte
	mask          []byte
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

// CheckMatch reads the file from the reader and checks if the signature matches.
//
// This function is designed to minimize the number of read syscalls by reading
// the file in chunks of "buffSize" bytes.
func (s *Signature) CheckMatch(r io.Reader) (*SigMatches, error) {
	var (
		offset   = 0
		offsets  []int
		buffer   = NewRingBuffer(buffSize)
		maxBytes = buffSize - s.length()

		fileByte byte
	)

	if maxBytes < 0 {
		maxBytes = buffSize
	}

	// Read the entire buffer the first time.
	n, err := buffer.Read(r, buffSize)
	if err != nil {
		return nil, err
	}

	for n > 0 {
		for i := 0; i < n; i++ {
			// If there isn't enough bytes to match the pattern, break.
			if i+s.length() >= buffer.Size() {
				break
			}

			// If the byte at i doesn't match the first byte of the pattern, continue.
			fileByte = buffer.Get(i) & s.mask[0]
			if fileByte != s.maskedPattern[0] {
				continue
			}

			// Check if the rest of the pattern matches.
			for j := 1; j < s.length(); j++ {
				fileByte = buffer.Get(i+j) & s.mask[j]

				if s.maskedPattern[j] != fileByte {
					break
				}

				// If the last iteration of the loop is reached, the pattern matches.
				if j == s.length()-1 {
					offsets = append(offsets, i+offset)
				}
			}
		}

		offset += maxBytes

		// Read the next chunk of the file.
		// We read less bytes than the buffer size to account for possible matches
		// that span the buffer boundary.
		n, err = buffer.Read(r, maxBytes)

		if err == io.EOF {
			// TODO check if this is necessary
			break
		} else if err != nil {
			return nil, err
		}
	}

	return &SigMatches{Offsets: offsets, Signature: s}, nil
}
