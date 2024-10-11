package signature

import (
	"fmt"
	"io"
)

type SigMatches struct {
	Signature *Signature
	IsMatch   bool
	Offsets   map[string]matchOffsets
}

func (sm *SigMatches) Len() int {
	return len(sm.Offsets)
}

func (sm *SigMatches) Write(w io.StringWriter, fileName string) {
	w.WriteString("================================================================================\n")
	w.WriteString(fmt.Sprintf("File:         %s\n", fileName))
	w.WriteString(fmt.Sprintf("Signature:    %s\n", sm.Signature.Name))
	w.WriteString(fmt.Sprintf("Description:  %s\n", sm.Signature.Description))
	w.WriteString("================================================================================\n")

	if !sm.IsMatch {
		w.WriteString("No matches found\n\n")
		return
	}

	w.WriteString(fmt.Sprintf("%d Matches found at offsets: \n", sm.Len()))
	// for i, offset := range sm.Offsets {
	// 	w.WriteString(
	// 		fmt.Sprintf(
	// 			"\t> [%d] offset=%d (dd if=%s bs=1 skip=%d count=%d 2>/dev/null | hexdump -C)\n",
	// 			i+1, offset, fileName, offset, sm.Signature.length(),
	// 		),
	// 	)
	// }

	w.WriteString("\n")
}
