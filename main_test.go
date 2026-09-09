package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLI(t *testing.T) {
	assert := assert.New(t)

	t.Run("Show help with -h", func(t *testing.T) {
		var (
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
			stdout = bytes.NewBuffer([]byte{})
			stderr = bytes.NewBuffer([]byte{})
		)

		exitCode, err := run([]string{}, stdout, stderr)

		assert.Nil(err)
		assert.Equal(exitFailure, exitCode)
		assert.Zero(stdout.Available())
		assert.Contains(string(stderr.Bytes()), "Usage: binmat [options] <file|directory>")
	})
}
