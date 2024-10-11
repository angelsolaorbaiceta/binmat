package signature

// A Signature is a pattern that can be matched in a file.
// A Signature is defined by a name, a description, a pattern, and a mask.
// The pattern is the sequence of bytes that must be matched.
// The mask is applied to the pattern to define which bytes must be matched, and which can be ignored.
type Signature struct {
	Name        string
	Description string
	patterns    map[string]*signaturePattern
	condition   string
}

// Make creates a new Signature with the given name, description, patterns,
// and condition.
func Make(
	name, description string,
	patterns map[string]*signaturePattern,
	condition string,
) *Signature {
	// TODO: validate that all names in the condition are in the patterns.
	return &Signature{
		Name:        name,
		Description: description,
		patterns:    patterns,
		condition:   condition,
	}
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

	matches := make(map[string]matchOffsets)
	for range s.patterns {
		match := <-ch
		matches[match.name] = match.matches
	}

	// TODO: asses condition

	return &SigMatches{
		IsMatch:   true,
		Signature: s,
		Offsets:   matches,
	}
}
