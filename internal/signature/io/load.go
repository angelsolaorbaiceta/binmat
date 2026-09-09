package io

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/angelsolaorbaiceta/binmat/internal/signature"
)

// LoadSignatures loads the signatures from the .yaml files found at the given
// directory path, typically "$HOME/.config/binmat".
func LoadSignatures(path string) (signature.Signatures, error) {
	yamlFilePaths, err := findYamlFiles(path)
	if err != nil {
		return nil, err
	}

	signatures := make(signature.Signatures, 0, len(yamlFilePaths))

	for _, filePath := range yamlFilePaths {
		sig, err := loadSignature(filePath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filePath, err)
		}

		signatures = append(signatures, sig)
	}

	return signatures, nil
}

func loadSignature(filePath string) (signature.Signature, error) {
	r, err := os.Open(filePath)
	if err != nil {
		return signature.Signature{}, err
	}
	defer r.Close()

	return signature.ReadFromYaml(r)
}

// findYamlFiles returns a slice of full paths to all .yaml files found in the
// passed in directory. Directories aren't recursively explored, just the top
// level is searched.
func findYamlFiles(path string) ([]string, error) {
	var yamlFiles []string

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			yamlFiles = append(yamlFiles, filepath.Join(path, entry.Name()))
		}
	}

	return yamlFiles, nil
}
