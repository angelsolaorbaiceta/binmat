package signature

import (
	"fmt"
	"io"
	"os"
)

type SigMatchMeta struct {
	FilePath string
}

// A SigMatch is the result of attempting to match a file against a signature.
type SigMatch struct {
	Meta      SigMatchMeta
	Signature *Signature
	IsMatch   bool
	Offsets   map[string]matchOffsets
}

func (sm *SigMatch) Len() int {
	return len(sm.Offsets)
}

func (sm *SigMatch) Write(w io.StringWriter) {
	w.WriteString("================================================================================\n")
	w.WriteString(fmt.Sprintf("File:         %s\n", sm.Meta.FilePath))
	w.WriteString(fmt.Sprintf("Signature:    %s\n", sm.Signature.Name))
	w.WriteString(fmt.Sprintf("Description:  %s\n", sm.Signature.Description))
	w.WriteString("================================================================================\n")

	if !sm.IsMatch {
		w.WriteString("No matches found\n\n")
		return
	}

	w.WriteString(fmt.Sprintf("%d Matches found at offsets: \n", sm.Len()))
	w.WriteString("\n")
}

func SearchMatches(sigs Signatures, path string) []SigMatch {
	var (
		isDir   bool
		matches []SigMatch
		err     error
	)

	if stat, err := os.Stat(path); err != nil {
		fmt.Fprintf(os.Stderr, "Can't get '%s' file info: %s\n", path, err)
		os.Exit(1)
	} else {
		isDir = stat.IsDir()
	}

	if isDir {
		matches, err = sigs.CheckDir(path)
	} else {
		matches, err = sigs.Check(path)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't check for matches: %s\n", err)
		os.Exit(1)
	}

	return matches
}
