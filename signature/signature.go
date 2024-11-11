package signature

import "github.com/angelsolaorbaiceta/binmat/bexpr"

// A Signature is a pattern that can be matched in a file.
// A Signature is defined by a name, a description, a pattern, and a mask.
// The pattern is the sequence of bytes that must be matched.
// The mask is applied to the pattern to define which bytes must be matched, and which can be ignored.
type Signature struct {
	Name        string
	Description string
	patterns    map[string]*signaturePattern
	condition   string
	conditionFn bexpr.Condition
}

// Make creates a new Signature with the given name, description, patterns,
// and condition.
// If the condition can't be successfully parsed, an error is returned.
func Make(
	name, description string,
	patterns map[string]*signaturePattern,
	condition string,
) (*Signature, error) {
	conditionFn, err := bexpr.ParseCondition(condition)
	if err != nil {
		// TODO: use a signature domain error
		return nil, err
	}
	// TODO: validate that all names in the condition are in the patterns.
	return &Signature{
		Name:        name,
		Description: description,
		patterns:    patterns,
		condition:   condition,
		conditionFn: conditionFn,
	}, nil
}

// length returns the number of patterns in the signature.
func (s *Signature) length() int {
	return len(s.patterns)
}

// CheckMatch reads the file from the byte slice and checks each of the patterns
// in the signature in parallel. It returns a SigMatches struct with the results.
//
// The function expects the full file contents in a byte slice, as binaries themselves
// are usually small enough to fit in memory.
func (s *Signature) CheckMatch(data []byte) *SigMatches {
	ch := make(chan struct {
		name    string
		matches matchOffsets
	})

	for name, pattern := range s.patterns {
		go func(name string, pattern *signaturePattern) {
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
	for range s.patterns {
		match := <-ch
		matchOffs[match.name] = match.matches
		matchVars[match.name] = match.matches.isMatch()
	}

	// TODO: handle error
	isMatch, _ := s.conditionFn(matchVars)

	return &SigMatches{
		IsMatch:   isMatch,
		Signature: s,
		Offsets:   matchOffs,
	}
}
