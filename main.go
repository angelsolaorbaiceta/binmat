package main

import (
	"fmt"
	"os"
	"path/filepath"

	sigio "github.com/angelsolaorbaiceta/binmat/signature/io"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file|directory>\n", os.Args[0])
		os.Exit(1)
	}

	homePath, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting the user's home path: %s\n", err)
		os.Exit(1)
	}

	sigsPath := filepath.Join(homePath, ".config/binmat")
	sigs, err := sigio.LoadSignatures(sigsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading the .yaml signatures from '%s': %s\n", sigsPath, err)
		os.Exit(1)
	}

	matches, _ := sigs.SearchMatches(os.Args[1])
	fmt.Printf("Scanned %d files.\n", len(matches))
	for _, match := range matches {
		if match.IsMatch {
			match.Write(os.Stdout)
		}
	}
}
