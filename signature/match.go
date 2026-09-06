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
