package signature

import (
	"io"
)

const (
	NoMask     = 0xff
	IgnoreMask = 0x00
)

type Signature struct {
	Name        string
	Description string
	Pattern     []byte
	Mask        []byte
}

// Make creates a new Signature with the given name, description, and pattern.
func Make(name, description string, pattern []byte) *Signature {
	mask := make([]byte, len(pattern))
	for i := range mask {
		mask[i] = NoMask
	}

	return &Signature{
		Name:        name,
		Description: description,
		Pattern:     pattern,
		Mask:        mask,
	}
}

// MakeWithMask creates a new Signature with the given name, description,
// pattern, and mask. The pattern and mask must be the same length.
func MakeWithMask(name, description string, pattern, mask []byte) *Signature {
	if len(pattern) != len(mask) {
		panic("pattern and mask length mismatch")
	}

	return &Signature{
		Name:        name,
		Description: description,
		Pattern:     pattern,
		Mask:        mask,
	}
}

func (s *Signature) patternLen() int {
	return len(s.Pattern)
}

func (s *Signature) CheckMatch(r io.Reader) ([]int, error) {
	var (
		offset  = 0
		offsets []int
		ok      bool
		stride  int
		buffer  = NewRingBuffer(s.patternLen())
	)

	n, err := buffer.Read(r, s.patternLen())
	if err != nil {
		return nil, err
	}
	if n < s.patternLen() {
		// No enough bytes to match, return no offsets.
		return []int{}, nil
	}

	if ok, stride = s.checkMatch(buffer); ok {
		offsets = append(offsets, offset)
	}

	offset += stride

	// Read the rest of the file, "stride" bytes at a time.
	for {
		n, err := buffer.Read(r, stride)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if n == 0 {
			break
		}

		if ok, stride = s.checkMatch(buffer); ok {
			offsets = append(offsets, offset)
		}
		offset += stride
	}

	return offsets, nil
}

// checkMatch checks that the passed bytes match the pattern and mask
// of the signature, expecting the bytes to be the same length as the
// pattern and mask.
//
// It returns true if the bytes match the pattern and mask, and false otherwise.
// It also returns the stride: the number of bytes to move the buffer forward.
// The stride can be the entire length of the pattern if there is a match or
// there is no possible match inside the buffer, or the number of bytes where
// the next possible match could start.
func (s *Signature) checkMatch(buffer ByteBuffer) (bool, int) {
	var (
		mask      byte
		pattern   byte
		buffByte  byte
		firstByte = s.Pattern[0] & s.Mask[0]
		stride    = 0
		matches   = true
	)

	if buffer.Size() != s.patternLen() {
		panic("buffer size does not match pattern length")
	}

	for i := 0; i < s.patternLen(); i++ {
		mask = s.Mask[i]
		pattern = s.Pattern[i]
		buffByte = buffer.Get(i)

		// If the byte matches, use it as the stride.
		// The stride should only be set once to be > 0, the first time a match is found.
		if (firstByte == (mask & buffByte)) && stride == 0 {
			stride = i
		}

		if (mask & pattern) != (mask & buffByte) {
			matches = false

			// If a stride is found, return false and the stride.
			// If the stride is 0, continue iterating to find a match.
			if stride > 0 {
				return false, stride
			}
		}
	}

	if matches {
		return true, s.patternLen()
	} else {
		// If no stride was found, use the length of the pattern.
		if stride == 0 {
			stride = s.patternLen()
		}

		return false, stride
	}
}
