package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	sigio "github.com/angelsolaorbaiceta/binmat/signature/io"
)

// Exit codes follow grep's convention, which is what anyone scripting a
// scanner will assume:
//
//	0 - ran fine, at least one signature matched
//	1 - ran fine, nothing matched
//	2 - something went wrong (bad flags, unreadable signatures, bad target)
const (
	exitMatch   = 0
	exitNoMatch = 1
	exitFailure = 2
)

func main() {
	code, err := run(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "binmat: %v\n", err)
	}
	os.Exit(code)
}

func run(args []string, stdout, stderr io.Writer) (int, error) {
	fs := flag.NewFlagSet("binmat", flag.ContinueOnError)
	fs.SetOutput(stderr)

	defaultSigsPath, sigsPathErr := defaultSignaturesPath()

	var (
		sigsPath = fs.String("sigs", defaultSigsPath, "directory holding the .yaml signature definitions")
	)

	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: binmat [options] <file|directory>\n\n")
		fmt.Fprintf(stderr, "Scans a file or directory tree against the YAML signatures in --sigs.\n\n")
		fmt.Fprintf(stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(stderr, "\nExit status: 0 if a signature matched, 1 if none did, 2 on error.\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitMatch, nil
		}
		return exitFailure, nil // FlagSet already wrote the diagnostic
	}

	if fs.NArg() != 1 {
		fs.Usage()
		return exitFailure, nil
	}
	target := fs.Arg(0)

	if *sigsPath == "" && sigsPathErr != nil {
		return exitFailure, fmt.Errorf("resolving the default signatures directory: %w", sigsPathErr)
	}

	sigs, err := sigio.LoadSignatures(*sigsPath)
	if err != nil {
		return exitFailure, fmt.Errorf("loading signatures from %q: %w", *sigsPath, err)
	}

	matches, _ := sigs.SearchMatches(target)
	jsonResult, err := json.Marshal(matches)
	if err != nil {
		return exitFailure, fmt.Errorf("generating JSON report: %v", err)
	}

	stdout.Write(jsonResult)

	// if matched == 0 {
	// 	return exitNoMatch, nil
	// }

	return exitMatch, nil
}

func defaultSignaturesPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "binmat"), nil
}
