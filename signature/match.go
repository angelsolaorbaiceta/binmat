package signature

import (
	"fmt"
	"io"
)

type SigMatchMeta struct {
	FilePath string
}

// A SigMatch is the result of attempting to match a file against a signature.
type SigMatch struct {
	Meta      *SigMatchMeta
	Signature *Signature
	IsMatch   bool
	Offsets   map[string]matchOffsets
}

func (sm *SigMatch) Len() int {
	return len(sm.Offsets)
}

func (sm *SigMatch) Write(w io.Writer) {
	io.WriteString(w, "================================================================================\n")
	io.WriteString(w, fmt.Sprintf("File:         %s\n", sm.Meta.FilePath))
	io.WriteString(w, fmt.Sprintf("Signature:    %s\n", sm.Signature.Name))
	io.WriteString(w, fmt.Sprintf("Description:  %s\n", sm.Signature.Description))
	io.WriteString(w, "================================================================================\n")

	if !sm.IsMatch {
		io.WriteString(w, "No matches found\n\n")
		return
	}

	io.WriteString(w, fmt.Sprintf("%d Matches found at offsets: \n", sm.Len()))
	io.WriteString(w, "\n")
}
