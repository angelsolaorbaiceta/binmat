package signature

import (
	"encoding/json"
	"io"
)

// A SigMatch is the result of attempting to match a file against a signature.
type SigMatch struct {
	FilePath      string `json:"filePath"`
	SignatureName string `json:"signatureName"`
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
