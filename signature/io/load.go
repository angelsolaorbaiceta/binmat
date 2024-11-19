package io

import (
	"os"
	"path/filepath"

	"github.com/angelsolaorbaiceta/binmat/signature"
)

// LoadSignatures loads the signatures from the .yaml files found at the given
// directory path, typically "$HOME/.config/binmat".
func LoadSignatures(path string) (signature.Signatures, error) {
	signatures, err := loadIOSignatures(path)
	if err != nil {
		return nil, err
	}

	domainSigs, err := signaturesToDomain(signatures)
	if err != nil {
		return nil, err
	}

	return domainSigs, nil
}

func loadIOSignatures(path string) ([]Signature, error) {
	yamlFilePaths, err := findYamlFiles(path)
	if err != nil {
		return nil, err
	}

	signatures := make([]Signature, len(yamlFilePaths))

	for i, filePath := range yamlFilePaths {
		r, err := os.Open(filePath)
		if err != nil {
			return signatures, err
		}

		sig, err := ReadFromYaml(r)
		if err != nil {
			return signatures, err
		}

		signatures[i] = sig
	}

	return signatures, nil
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
