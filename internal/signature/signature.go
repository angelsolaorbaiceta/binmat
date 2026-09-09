package signature

import (
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/angelsolaorbaiceta/binmat/internal/bexpr"
)

// A SigMatch is the result of attempting to match a file against a signature.
type SigMatch struct {
	FilePath           string `json:"filePath"`
	SignatureName      string `json:"signatureName"`
	SignatureCondition string `json:"signatureCondition"`
	// Whether the signature condition was met.
	IsMatch bool `json:"isMatch"`
	// The offsets at which each signature pattern was found in the file.
	OffsetsByPattern map[string]PatternMatchOffsets `json:"offsetsByPattern"`
}

func (sm *SigMatch) Len() int {
	return len(sm.OffsetsByPattern)
}

func (sm *SigMatch) WriteJSON(w io.Writer) error {
	jsonData, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	w.Write(jsonData)
	return nil
}

// A Signature is a pattern that can be matched in a file.
// A Signature is defined by a name, a description, a pattern, and a mask.
// The pattern is the sequence of bytes that must be matched.
// The mask is applied to the pattern to define which bytes must be matched, and
// which can be ignored.
type Signature struct {
	Name        string
	Description string
	Patterns    map[string]*SignaturePattern
	Condition   string
	conditionFn bexpr.Condition
}

// Make creates a new Signature with the given name, description, patterns,
// and condition.
// If the condition can't be successfully parsed, an error is returned.
// If any of the pattern names doesn't adhere to the convention, an error is returned.
func Make(
	name, description string,
	patterns map[string]*SignaturePattern,
	condition string,
) (Signature, error) {
	var signature Signature

	if len(strings.TrimSpace(name)) == 0 {
		return signature, ErrSignature{reason: ErrSigEmptyName}
	}

	if len(patterns) == 0 {
		return signature, ErrSignature{reason: ErrSigEmptyPatterns}
	}

	if len(strings.TrimSpace(condition)) == 0 {
		return signature, ErrSignature{reason: ErrSigWrongCondition}
	}

	conditionFn, err := bexpr.ParseCondition(condition)
	if err != nil {
		return signature, ErrSignature{reason: ErrSigWrongCondition, cause: err}
	}

	// Create a map where all pattern names are assigned "true" to test if the
	// conditionFn has all the variables it needs.
	varsMap := make(map[string]bool)
	for name := range patterns {
		varsMap[name] = true
	}

	if _, err := conditionFn(varsMap); err != nil {
		return signature, ErrSignature{reason: ErrSigMissingPattern, cause: err}
	}

	signature.Name = name
	signature.Description = description
	signature.Patterns = patterns
	signature.Condition = condition
	signature.conditionFn = conditionFn

	return signature, nil
}

// CheckMatch reads the file from the byte slice and checks each of the patterns
// in the signature in parallel. It returns a SigMatches struct with the results.
func (s Signature) CheckMatch(data []byte, filePath string) *SigMatch {
	ch := make(chan struct {
		name    string
		matches PatternMatchOffsets
	})

	for name, pattern := range s.Patterns {
		go func(name string, pattern *SignaturePattern) {
			ch <- struct {
				name    string
				matches PatternMatchOffsets
			}{
				name:    name,
				matches: pattern.checkMatch(data),
			}
		}(name, pattern)
	}

	var (
		matchOffs = make(map[string]PatternMatchOffsets)
		matchVars = make(map[string]bool)
	)
	for range s.Patterns {
		match := <-ch
		matchOffs[match.name] = match.matches
		matchVars[match.name] = match.matches.isMatch()
	}

	// All the variables names (patterns) in the condition have been checked to
	// be present in the patterns map. No error should be returned here.
	isMatch, _ := s.conditionFn(matchVars)

	return &SigMatch{
		FilePath:           filePath,
		SignatureName:      s.Name,
		SignatureCondition: s.Condition,
		IsMatch:            isMatch,
		OffsetsByPattern:   matchOffs,
	}
}

// Signatures is a collection of byte Signatures.
type Signatures []Signature

// MatchesBySignatureName groups the match results of every scanned file under
// the name of the signature they were checked against.
type MatchesBySignatureName map[string][]*SigMatch

// add appends the given matches to the entry of their signature.
func (m MatchesBySignatureName) add(matches []*SigMatch) {
	for _, match := range matches {
		if match.IsMatch {
			m[match.SignatureName] = append(m[match.SignatureName], match)
		}
	}
}

// SearchMatches walks the given root:
//   - if root is a file, it checks that file.
//   - if root is a directory, it recursively checks every file.
//
// It collects all matches and returns them, along with any errors encountered.
func (s Signatures) SearchMatches(root string) (MatchesBySignatureName, []error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, []error{err}
	}

	result := make(MatchesBySignatureName)

	if !info.IsDir() {
		matches, err := s.checkFile(root)
		if err != nil {
			return result, []error{err}
		}
		result.add(matches)

		return result, nil
	}

	var errs []error

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// WalkDir itself encountered an issue (e.g., permission denied).
			// We record the error but continue walking the rest.
			errs = append(errs, err)
			return nil // return nil to keep walking
		}

		if d.IsDir() {
			return nil // skip directories
		}

		matches, err := s.checkFile(path)
		if err != nil {
			errs = append(errs, err)
			return nil // continue with the next file
		}

		result.add(matches)
		return nil
	})

	// If WalkDir itself returns a fatal error (e.g., root doesn't exist),
	// we return it; otherwise we return the collected errors.
	if err != nil && len(result) == 0 && len(errs) == 0 {
		return nil, []error{err}
	}
	return result, errs
}

func (s Signatures) checkFile(binPath string) ([]*SigMatch, error) {
	data, err := readFileBytes(binPath)
	if err != nil {
		return nil, err
	}

	matches := make([]*SigMatch, len(s))
	for i, sig := range s {
		match := sig.CheckMatch(data, binPath)
		matches[i] = match
	}

	return matches, nil
}
