package signature

import (
	"io"
	"os"
)

// Signatures is a collection of byte Signatures.
type Signatures []*Signature

// Check reads the file from the byte slice and checks if the signatures match.
// It returns all the matches found, or an error if there is a problem reading the file.
func (s Signatures) Check(binPath string) ([]*SigMatches, error) {
	var (
		results = make(chan *SigMatches)
		matches []*SigMatches
	)

	reader, err := os.Open(os.Args[1])
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	for _, sig := range s {
		go func(sig *Signature) {
			results <- sig.CheckMatch(data)
		}(sig)
	}

	for range s {
		match := <-results
		if match.Len() > 0 {
			matches = append(matches, match)
		}
	}

	return matches, nil
}
