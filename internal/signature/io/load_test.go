package io

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadSignatures(t *testing.T) {
	fixture, err := os.ReadFile("../../../examples/signatures/__io_test.yaml")
	if err != nil {
		panic("Can't read file" + err.Error())
	}

	t.Run("loads only the top-level .yaml files", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "one.yaml"), fixture, 0o644)
		os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("not a signature"), 0o644)
		os.MkdirAll(filepath.Join(dir, "nested"), 0o755)
		os.WriteFile(filepath.Join(dir, "nested", "two.yaml"), fixture, 0o644)

		sigs, err := LoadSignatures(dir)

		assert.Nil(t, err)
		assert.Len(t, sigs, 1)
		assert.Equal(t, "A test signature", sigs[0].Name)
	})

	t.Run("fails if any signature is invalid", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "good.yaml"), fixture, 0o644)
		os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("name: bad\npatterns:\n  a: '{ 01 b }'\ncondition: a\n"), 0o644)

		sigs, err := LoadSignatures(dir)

		assert.Nil(t, sigs)
		assert.ErrorContains(t, err, "bad.yaml")
	})

	t.Run("fails on a missing directory", func(t *testing.T) {
		_, err := LoadSignatures(filepath.Join(t.TempDir(), "nope"))

		assert.NotNil(t, err)
	})
}
