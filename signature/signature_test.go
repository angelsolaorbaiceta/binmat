package signature

import (
	"bytes"
	"testing"
)

func TestMatchSignatureWithoutMask(t *testing.T) {
	sig := Make("test", "test signature", []byte{0x01, 0x02, 0x03})

	t.Run("No enough bytes to match", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0x02})
			matches, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if matches.Len() != 0 {
			t.Fatalf("expected 0 offsets, got %d", matches.Len())
		}
	})

	t.Run("One match starting at offset 0", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
			matches, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if matches.Len() != 1 {
			t.Fatalf("expected 1 offset, got %d", matches.Len())
		}
		if matches.Offsets[0] != 0 {
			t.Fatalf("expected offset 0, got %d", matches.Offsets[0])
		}
	})

	t.Run("One match starting at offset 2", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05})
			matches, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if matches.Len() != 1 {
			t.Fatalf("expected 1 offset, got %d", matches.Len())
		}
		if matches.Offsets[0] != 2 {
			t.Fatalf("expected offset 2, got %d", matches.Offsets[0])
		}
	})

	t.Run("No match", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0x02, 0x04, 0x05, 0x06, 0x07})
			matches, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if matches.Len() != 0 {
			t.Fatalf("expected 0 offsets, got %d", matches.Len())
		}
	})

	t.Run("Multiple matches", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05})
			matches, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if matches.Len() != 2 {
			t.Fatalf("expected 2 offsets, got %d", matches.Len())
		}
		if matches.Offsets[0] != 0 {
			t.Fatalf("expected offset 0, got %d", matches.Offsets[0])
		}
		if matches.Offsets[1] != 5 {
			t.Fatalf("expected offset 5, got %d", matches.Offsets[1])
		}
	})

	t.Run("Matches at the buffer's boundaries", func(t *testing.T) {
		buffSize = 5

		var (
			reader       = bytes.NewReader([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05})
			matches, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if matches.Len() != 2 {
			t.Fatalf("expected 2 offsets, got %d", matches.Len())
		}
		if matches.Offsets[0] != 0 {
			t.Fatalf("expected offset 0, got %d", matches.Offsets[0])
		}
		if matches.Offsets[1] != 5 {
			t.Fatalf("expected offset 5, got %d", matches.Offsets[1])
		}
	})
}

func TestMatchSignatureWithMask(t *testing.T) {
	sig := MakeWithMask(
		"test", "test signature",
		[]byte{0x01, 0x02, 0x03},
		[]byte{MatchByte, AnyByte, MatchByte},
	)

	t.Run("No enough bytes to match", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0x02})
			matches, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if matches.Len() != 0 {
			t.Fatalf("expected 0 offsets, got %d", matches.Len())
		}
	})

	t.Run("One match starting at offset 0", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0xab, 0x03, 0x04, 0x05})
			matches, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if matches.Len() != 1 {
			t.Fatalf("expected 1 offset, got %d", matches.Len())
		}
		if matches.Offsets[0] != 0 {
			t.Fatalf("expected offset 0, got %d", matches.Offsets[0])
		}
	})

	t.Run("One match starting at offset 2", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x00, 0x00, 0x01, 0xab, 0x03, 0x04, 0x05})
			matches, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if matches.Len() != 1 {
			t.Fatalf("expected 1 offset, got %d", matches.Len())
		}
		if matches.Offsets[0] != 2 {
			t.Fatalf("expected offset 2, got %d", matches.Offsets[0])
		}
	})

	t.Run("No match", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0xab, 0x04, 0x05, 0x06, 0x07})
			matches, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if matches.Len() != 0 {
			t.Fatalf("expected 0 offsets, got %d", matches.Len())
		}
	})

	t.Run("Multiple matches", func(t *testing.T) {
		var (
			reader       = bytes.NewReader([]byte{0x01, 0xab, 0x03, 0x04, 0x05, 0x01, 0xcd, 0x03, 0x04, 0x05})
			matches, err = sig.CheckMatch(reader)
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if matches.Len() != 2 {
			t.Fatalf("expected 2 offsets, got %d", matches.Len())
		}
		if matches.Offsets[0] != 0 {
			t.Fatalf("expected offset 0, got %d", matches.Offsets[0])
		}
		if matches.Offsets[1] != 5 {
			t.Fatalf("expected offset 5, got %d", matches.Offsets[1])
		}
	})
}
