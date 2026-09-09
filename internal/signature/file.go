package signature

import (
	"io"
	"os"
)

func readFileBytes(filePath string) ([]byte, error) {
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return data, nil
}
