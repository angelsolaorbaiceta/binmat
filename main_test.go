package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// sigAYaml matches when the header is found and then either the marker string
// is present or the kill switch is absent:
//
//	header AND (marker OR NOT kill)
//
// The header's third byte is a wildcard, so "4d 5a 90 00" and "4d 5a 00 00"
// both match.
const sigAYaml = `name: Signature A
description: An MZ header with a wildcard byte, plus a marker string or no kill switch
patterns:
  header: '{ 4d 5a ?? 00 }'
  marker: binmat
  kill: '{ de ad be ef }'
condition: header AND (marker OR NOT kill)
`

// sigBYaml matches a little-endian ELF that is neither a debug build nor
// UPX-packed:
//
//	elf AND NOT (debug OR packed)
//
// The ELF pattern wildcards the class byte (32/64-bit) but pins the
// endianness byte to little endian, so "7f 45 4c 46 01 01" and
// "7f 45 4c 46 02 01" match while "7f 45 4c 46 02 02" doesn't.
const sigBYaml = `name: Signature B
description: A little-endian ELF that is neither a debug build nor UPX-packed
patterns:
  elf: '{ 7f 45 4c 46 ?? 01 }'
  debug: debug build
  packed: UPX!
condition: elf AND NOT (debug OR packed)
`

func TestCLI(t *testing.T) {
	sigsDir := filepath.Join(t.TempDir(), "sigs")
	if err := os.Mkdir(sigsDir, 0o755); err != nil {
		t.Fatalf("creating the signatures directory: %v", err)
	}

	for name, doc := range map[string]string{"a.yaml": sigAYaml, "b.yaml": sigBYaml} {
		if err := os.WriteFile(filepath.Join(sigsDir, name), []byte(doc), 0o644); err != nil {
			t.Fatalf("writing the %s signature: %v", name, err)
		}
	}

	t.Run("Show help with -h", func(t *testing.T) {
		var (
			assert = assert.New(t)
			stdout = bytes.NewBuffer([]byte{})
			stderr = bytes.NewBuffer([]byte{})
		)

		exitCode, err := run([]string{"-h"}, stdout, stderr)

		assert.Nil(err)
		assert.Equal(exitMatch, exitCode)
		assert.Zero(stdout.Available())
		assert.Contains(string(stderr.Bytes()), "Usage: binmat [options] <file|directory>")
	})

	t.Run("Missing path argument", func(t *testing.T) {
		var (
			assert = assert.New(t)
			stdout = bytes.NewBuffer([]byte{})
			stderr = bytes.NewBuffer([]byte{})
		)

		exitCode, err := run([]string{}, stdout, stderr)

		assert.Nil(err)
		assert.Equal(exitFailure, exitCode)
		assert.Zero(stdout.Available())
		assert.Contains(string(stderr.Bytes()), "Usage: binmat [options] <file|directory>")
	})

	t.Run("Matches a file against the signatures", func(t *testing.T) {
		var (
			assert = assert.New(t)
			stdout = bytes.NewBuffer([]byte{})
			stderr = bytes.NewBuffer([]byte{})
			target = filepath.Join(t.TempDir(), "target.bin")
		)
		// Signature A: header (wildcard byte set to 0x90) followed by the marker.
		// Signature B: the ELF magic isn't present.
		os.WriteFile(target, append([]byte{0x4d, 0x5a, 0x90, 0x00}, "binmat"...), 0o644)

		exitCode, err := run([]string{"-sigs", sigsDir, target}, stdout, stderr)

		// The path is JSON-encoded so it escapes the same way the CLI does.
		// A pattern with no hits is reported as null (a nil offsets slice).
		wantPath, _ := json.Marshal(target)
		want := fmt.Sprintf(`{
			"Signature A": [{
				"filePath": %s,
				"signatureName": "Signature A",
				"signatureCondition": "header AND (marker OR NOT kill)",
				"isMatch": true,
				"offsetsByPattern": {"header": [0], "marker": [4], "kill": null}
			}]
		}`, wantPath)

		assert.Nil(err)
		assert.Equal(exitMatch, exitCode)
		assert.Zero(stderr.Len())
		assert.JSONEq(want, stdout.String())
	})

	t.Run("Reports no match", func(t *testing.T) {
		var (
			assert = assert.New(t)
			stdout = bytes.NewBuffer([]byte{})
			stderr = bytes.NewBuffer([]byte{})
			target = filepath.Join(t.TempDir(), "target.bin")
		)
		os.WriteFile(target, []byte("nothing to see here"), 0o644)

		exitCode, err := run([]string{"-sigs", sigsDir, target}, stdout, stderr)

		assert.Nil(err)
		assert.Equal(exitNoMatch, exitCode)
		assert.Zero(stderr.Len())
		assert.Equal("{}", stdout.String())
	})
}
