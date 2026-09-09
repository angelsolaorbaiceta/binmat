package signature

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/angelsolaorbaiceta/binmat/internal/bexpr"
	"gopkg.in/yaml.v3"
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

// A Signature identifies a binary through a set of named patterns and a
// boolean condition over them. It is both the YAML document users write and
// the compiled form used for matching: the exported fields hold the textual
// definition, the unexported ones hold their compiled counterparts.
type Signature struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Patterns    map[string]string `yaml:"patterns"`
	Condition   string            `yaml:"condition"`

	// patterns holds the parsed Patterns, by name.
	patterns map[string]*SignaturePattern
	// conditionFn holds the parsed Condition.
	conditionFn bexpr.Condition
}

// Make creates a new Signature with the given name, description, patterns,
// and condition, and compiles it so it's ready to match.
// If the condition or any of the patterns can't be parsed, or the condition
// refers to patterns that aren't defined, an ErrSignature is returned.
func Make(
	name, description string,
	patterns map[string]string,
	condition string,
) (Signature, error) {
	sig := Signature{
		Name:        name,
		Description: description,
		Patterns:    patterns,
		Condition:   condition,
	}

	if err := sig.compile(); err != nil {
		return Signature{}, err
	}

	return sig, nil
}

// ReadFromYaml decodes a Signature from its YAML representation.
// The returned signature is validated and compiled, ready to match.
func ReadFromYaml(r io.Reader) (Signature, error) {
	var sig Signature
	err := yaml.NewDecoder(r).Decode(&sig)

	return sig, err
}

// WriteYAML encodes the signature definition as YAML.
func (s Signature) WriteYAML(w io.Writer) error {
	enc := yaml.NewEncoder(w)
	if err := enc.Encode(s); err != nil {
		return err
	}

	return enc.Close()
}

// UnmarshalYAML decodes the textual definition and compiles it, so a
// Signature decoded from YAML is never left in an unusable state.
func (s *Signature) UnmarshalYAML(node *yaml.Node) error {
	// Decode into a type without this method so the decoder doesn't recurse.
	type plain Signature
	var decoded plain

	if err := node.Decode(&decoded); err != nil {
		return err
	}

	*s = Signature(decoded)

	return s.compile()
}

// compile validates the textual definition and parses the patterns and the
// condition into their runnable forms.
func (s *Signature) compile() error {
	if len(strings.TrimSpace(s.Name)) == 0 {
		return ErrSignature{reason: ErrSigEmptyName}
	}

	if len(s.Patterns) == 0 {
		return ErrSignature{reason: ErrSigEmptyPatterns}
	}

	if len(strings.TrimSpace(s.Condition)) == 0 {
		return ErrSignature{reason: ErrSigWrongCondition}
	}

	patterns := make(map[string]*SignaturePattern, len(s.Patterns))
	for name, raw := range s.Patterns {
		pattern, err := ParsePattern(raw)
		if err != nil {
			return ErrSignature{
				reason: ErrSigWrongPattern,
				cause:  fmt.Errorf("pattern %q: %w", name, err),
			}
		}

		patterns[name] = pattern
	}

	conditionFn, err := bexpr.ParseCondition(s.Condition)
	if err != nil {
		return ErrSignature{reason: ErrSigWrongCondition, cause: err}
	}

	// Create a map where all pattern names are assigned "true" to test if the
	// conditionFn has all the variables it needs.
	varsMap := make(map[string]bool, len(patterns))
	for name := range patterns {
		varsMap[name] = true
	}

	if _, err := conditionFn(varsMap); err != nil {
		return ErrSignature{reason: ErrSigMissingPattern, cause: err}
	}

	s.patterns = patterns
	s.conditionFn = conditionFn

	return nil
}

// CheckMatch reads the file from the byte slice and checks each of the patterns
// in the signature in parallel. It returns a SigMatches struct with the results.
func (s Signature) CheckMatch(data []byte, filePath string) *SigMatch {
	ch := make(chan struct {
		name    string
		matches PatternMatchOffsets
	})

	for name, pattern := range s.patterns {
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
	for range s.patterns {
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

// addIfMatch appends the given matches to the entry of their signature if the
// result matched.
func (m MatchesBySignatureName) addIfMatch(matches []*SigMatch) {
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
		result.addIfMatch(matches)

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

		result.addIfMatch(matches)
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
