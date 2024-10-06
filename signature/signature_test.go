package signature

import (
	"bytes"
	"testing"
)

func TestCheckMatch(t *testing.T) {
	b := NewRingBuffer(5)
	b.AddAll(0x01, 0x02, 0x03, 0x04, 0x05)

	t.Run("Exact match", func(t *testing.T) {
		var (
			sig        = Make("test", "test signature", []byte{0x01, 0x02, 0x03, 0x04, 0x05})
			ok, stride = sig.checkMatch(b)
		)

		if !ok {
			t.Fatalf("expected match, got no match")
		}
		// The entire pattern matched, so the stride should be the length of the pattern.
		if stride != sig.patternLen() {
			t.Fatalf("expected stride %d, got %d", sig.patternLen(), stride)
		}
	})

	t.Run("No byte matches the first byte of the pattern", func(t *testing.T) {
		var (
			sig        = Make("test", "test signature", []byte{0x06, 0x07, 0x08, 0x09, 0x0a})
			ok, stride = sig.checkMatch(b)
		)

		if ok {
			t.Fatalf("expected no match, got match")
		}
		// Because there is no match, the stride should be the length of the pattern.
		if stride != sig.patternLen() {
			t.Fatalf("expected stride %d, got %d", sig.patternLen(), stride)
		}
	})

	t.Run("Only the first byte matches the pattern", func(t *testing.T) {
		var (
			sig        = Make("test", "test signature", []byte{0x01, 0x07, 0x08, 0x09, 0x0a})
			ok, stride = sig.checkMatch(b)
		)

		if ok {
			t.Fatalf("expected no match, got match")
		}
		// Because there is no possible match other than the entire pattern (and this didn't match),
		// the stride should be the length of the pattern.
		if stride != sig.patternLen() {
			t.Fatalf("expected stride %d, got %d", sig.patternLen(), stride)
		}
	})

	t.Run("Only the second byte matches the first byte of the pattern", func(t *testing.T) {
		var (
			sig        = Make("test", "test signature", []byte{0x02, 0x07, 0x08, 0x09, 0x0a})
			ok, stride = sig.checkMatch(b)
		)

		if ok {
			t.Fatalf("expected no match, got match")
		}
		// The byte at i = 1 could be the start of a match, so the stride should be 1.
		if stride != 1 {
			t.Fatalf("expected stride 1, got %d", stride)
		}
	})
}

func TestMatchSignatureWithoutMask(t *testing.T) {
	sig := Make("test", "test signature", []byte{0x01, 0x02, 0x03})

	t.Run("No enough bytes to match", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0x02})
			offsets, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(offsets) != 0 {
			t.Fatalf("expected 0 offsets, got %d", len(offsets))
		}
	})

	t.Run("One match starting at offset 0", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
			offsets, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(offsets) != 1 {
			t.Fatalf("expected 1 offset, got %d", len(offsets))
		}
		if offsets[0] != 0 {
			t.Fatalf("expected offset 0, got %d", offsets[0])
		}
	})

	t.Run("One match starting at offset 2", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05})
			offsets, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(offsets) != 1 {
			t.Fatalf("expected 1 offset, got %d", len(offsets))
		}
		if offsets[0] != 2 {
			t.Fatalf("expected offset 2, got %d", offsets[0])
		}
	})

	t.Run("No match", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0x02, 0x04, 0x05, 0x06, 0x07})
			offsets, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(offsets) != 0 {
			t.Fatalf("expected 0 offsets, got %d", len(offsets))
		}
	})

	t.Run("Multiple matches", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05})
			offsets, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(offsets) != 2 {
			t.Fatalf("expected 2 offsets, got %d", len(offsets))
		}
		if offsets[0] != 0 {
			t.Fatalf("expected offset 0, got %d", offsets[0])
		}
		if offsets[1] != 5 {
			t.Fatalf("expected offset 5, got %d", offsets[1])
		}
	})
}

func TestMatchSignatureWithMask(t *testing.T) {
	sig := MakeWithMask(
		"test", "test signature",
		[]byte{0x01, 0x02, 0x03},
		[]byte{NoMask, IgnoreMask, NoMask},
	)

	t.Run("No enough bytes to match", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0x02})
			offsets, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(offsets) != 0 {
			t.Fatalf("expected 0 offsets, got %d", len(offsets))
		}
	})

	t.Run("One match starting at offset 0", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0xab, 0x03, 0x04, 0x05})
			offsets, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(offsets) != 1 {
			t.Fatalf("expected 1 offset, got %d", len(offsets))
		}
		if offsets[0] != 0 {
			t.Fatalf("expected offset 0, got %d", offsets[0])
		}
	})

	t.Run("One match starting at offset 2", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x00, 0x00, 0x01, 0xab, 0x03, 0x04, 0x05})
			offsets, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(offsets) != 1 {
			t.Fatalf("expected 1 offset, got %d", len(offsets))
		}
		if offsets[0] != 2 {
			t.Fatalf("expected offset 2, got %d", offsets[0])
		}
	})

	t.Run("No match", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0xab, 0x04, 0x05, 0x06, 0x07})
			offsets, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(offsets) != 0 {
			t.Fatalf("expected 0 offsets, got %d", len(offsets))
		}
	})

	t.Run("Multiple matches", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0xab, 0x03, 0x04, 0x05, 0x01, 0xcd, 0x03, 0x04, 0x05})
			offsets, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(offsets) != 2 {
			t.Fatalf("expected 2 offsets, got %d", len(offsets))
		}
		if offsets[0] != 0 {
			t.Fatalf("expected offset 0, got %d", offsets[0])
		}
		if offsets[1] != 5 {
			t.Fatalf("expected offset 5, got %d", offsets[1])
		}
	})
}
