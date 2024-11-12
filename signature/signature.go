package signature

import (
	"strings"

	"github.com/angelsolaorbaiceta/binmat/bexpr"
)

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
) (Signature, *ErrSignature) {
	var signature Signature

	if len(strings.TrimSpace(name)) == 0 {
		return signature, &ErrSignature{reason: ErrSigEmptyName}
	}

	if len(patterns) == 0 {
		return signature, &ErrSignature{reason: ErrSigEmptyPatterns}
	}

	if len(strings.TrimSpace(condition)) == 0 {
		return signature, &ErrSignature{reason: ErrSigWrongCondition}
	}

	conditionFn, err := bexpr.ParseCondition(condition)
	if err != nil {
		return signature, &ErrSignature{reason: ErrSigWrongCondition, cause: err}
	}

	// Create a map where all pattern names are assigned "true" to test if the
	// conditionFn has all the variables it needs.
	varsMap := make(map[string]bool)
	for name := range patterns {
		varsMap[name] = true
	}

	if _, err := conditionFn(varsMap); err != nil {
		return signature, &ErrSignature{reason: ErrSigMissingPattern, cause: err}
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
//
// The function expects the full file contents in a byte slice, as binaries themselves
// are usually small enough to fit in memory.
func (s Signature) CheckMatch(data []byte) *SigMatches {
	ch := make(chan struct {
		name    string
		matches matchOffsets
	})

	for name, pattern := range s.Patterns {
		go func(name string, pattern *SignaturePattern) {
			ch <- struct {
				name    string
				matches matchOffsets
			}{
				name:    name,
				matches: pattern.checkMatch(data),
			}
		}(name, pattern)
	}

	var (
		matchOffs = make(map[string]matchOffsets)
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

	return &SigMatches{
		IsMatch:   isMatch,
		Signature: &s,
		Offsets:   matchOffs,
	}
}
