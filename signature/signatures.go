package signature

import (
	"io/fs"
	"path/filepath"
)

// Signatures is a collection of byte Signatures.
type Signatures []Signature

// Check reads the file from the byte slice and checks if the signatures match.
// It returns all the matches found, or an error if there is a problem reading the file.
func (s Signatures) Check(binPath string) ([]SigMatch, error) {
	var (
		results = make(chan SigMatch)
		matches []SigMatch
		match   SigMatch
	)

	data, err := readFileBytes(binPath)
	if err != nil {
		return nil, err
	}

	for _, sig := range s {
		go func(sig *Signature) {
			match = sig.CheckMatch(data)
			match.Meta = SigMatchMeta{FilePath: binPath}
			results <- match
		}(&sig)
	}

	for range s {
		match := <-results
		if match.Len() > 0 {
			matches = append(matches, match)
		}
	}

	return matches, nil
}

// CheckDir checks every file inside the directory for matches against these
// signatures.
func (s Signatures) CheckDir(dirPath string) ([]SigMatch, error) {
	var matches []SigMatch

	err := filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fileMatches, err := s.Check(path)
			if err != nil {
				return err
			}

			matches = append(matches, fileMatches...)
		}

		return nil
	})

	return matches, err
}
