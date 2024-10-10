package signature

import (
	"testing"
)

func TestMatchPatternWithoutMask(t *testing.T) {
	sig := makePattern([]byte{0x01, 0x02, 0x03})

	t.Run("No enough bytes to match", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0x02}
			matches = sig.checkMatch(data)
		)

		if matches.len() != 0 {
			t.Fatalf("expected 0 offsets, got %d", len(matches))
		}
	})

	t.Run("One match starting at offset 0", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0x02, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		if matches.len() != 1 {
			t.Fatalf("expected 1 offset, got %d", matches.len())
		}
		if matches[0] != 0 {
			t.Fatalf("expected offset 0, got %d", matches[0])
		}
	})

	t.Run("One match starting at offset 2", func(t *testing.T) {
		var (
			data    = []byte{0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		if matches.len() != 1 {
			t.Fatalf("expected 1 offset, got %d", matches.len())
		}
		if matches[0] != 2 {
			t.Fatalf("expected offset 2, got %d", matches[0])
		}
	})

	t.Run("No match", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0x02, 0x04, 0x05, 0x06, 0x07}
			matches = sig.checkMatch(data)
		)

		if matches.len() != 0 {
			t.Fatalf("expected 0 offsets, got %d", matches.len())
		}
	})

	t.Run("Multiple matches", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		if matches.len() != 2 {
			t.Fatalf("expected 2 offsets, got %d", matches.len())
		}
		if matches[0] != 0 {
			t.Fatalf("expected offset 0, got %d", matches[0])
		}
		if matches[1] != 5 {
			t.Fatalf("expected offset 5, got %d", matches[1])
		}
	})
}

func TestMatchPatternWithMask(t *testing.T) {
	sig := makePatternWithMask(
		[]byte{0x01, 0x02, 0x03},
		[]byte{matchByte, anyByte, matchByte},
	)

	t.Run("No enough bytes to match", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0x02}
			matches = sig.checkMatch(data)
		)

		if matches.len() != 0 {
			t.Fatalf("expected 0 offsets, got %d", matches.len())
		}
	})

	t.Run("One match starting at offset 0", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0xab, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		if matches.len() != 1 {
			t.Fatalf("expected 1 offset, got %d", matches.len())
		}
		if matches[0] != 0 {
			t.Fatalf("expected offset 0, got %d", matches[0])
		}
	})

	t.Run("One match starting at offset 2", func(t *testing.T) {
		var (
			data    = []byte{0x00, 0x00, 0x01, 0xab, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		if matches.len() != 1 {
			t.Fatalf("expected 1 offset, got %d", matches.len())
		}
		if matches[0] != 2 {
			t.Fatalf("expected offset 2, got %d", matches[0])
		}
	})

	t.Run("No match", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0xab, 0x04, 0x05, 0x06, 0x07}
			matches = sig.checkMatch(data)
		)

		if matches.len() != 0 {
			t.Fatalf("expected 0 offsets, got %d", matches.len())
		}
	})

	t.Run("Multiple matches", func(t *testing.T) {
		var (
			data    = []byte{0x01, 0xab, 0x03, 0x04, 0x05, 0x01, 0xcd, 0x03, 0x04, 0x05}
			matches = sig.checkMatch(data)
		)

		if matches.len() != 2 {
			t.Fatalf("expected 2 offsets, got %d", matches.len())
		}
		if matches[0] != 0 {
			t.Fatalf("expected offset 0, got %d", matches[0])
		}
		if matches[1] != 5 {
			t.Fatalf("expected offset 5, got %d", matches[1])
		}
	})
}
