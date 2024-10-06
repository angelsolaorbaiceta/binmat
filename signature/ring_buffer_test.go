package signature

import "testing"

func TestRingBuffer(t *testing.T) {
	t.Run("Create an empty ring buffer", func(t *testing.T) {
		buff := NewRingBuffer(3)

		if buff.Size() != 0 {
			t.Fatalf("expected size 0, got %d", buff.Size())
		}
		if buff.Capacity() != 3 {
			t.Fatalf("expected capacity 3, got %d", buff.Capacity())
		}
		if got := buff.Get(0); got != 0 {
			t.Fatalf("expected byte 0, got 0x%02x", got)
		}
	})

	t.Run("Add a byte to a ring buffer", func(t *testing.T) {
		buff := NewRingBuffer(3)
		buff.Add(0x01)

		if buff.Size() != 1 {
			t.Fatalf("expected size 1, got %d", buff.Size())
		}
		if got := buff.Get(0); got != 0x01 {
			t.Fatalf("expected byte 0x01, got 0x%02x", got)
		}
	})

	t.Run("Add multiple bytes to a ring buffer", func(t *testing.T) {
		buff := NewRingBuffer(3)
		buff.AddAll(0x01, 0x02, 0x03)

		if buff.Size() != 3 {
			t.Fatalf("expected size 3, got %d", buff.Size())
		}
		if got := buff.Get(0); got != 0x01 {
			t.Fatalf("expected byte 0x01, got 0x%02x", got)
		}
		if got := buff.Get(1); got != 0x02 {
			t.Fatalf("expected byte 0x02, got 0x%02x", got)
		}
		if got := buff.Get(2); got != 0x03 {
			t.Fatalf("expected byte 0x03, got 0x%02x", got)
		}
	})

	t.Run("Add a byte to a full ring buffer", func(t *testing.T) {
		buff := NewRingBuffer(3)
		buff.AddAll(0x01, 0x02, 0x03)
		buff.Add(0x04)

		if got := buff.Get(0); got != 0x02 {
			t.Fatalf("expected byte 0x02, got 0x%02x", got)
		}
		if got := buff.Get(1); got != 0x03 {
			t.Fatalf("expected byte 0x03, got 0x%02x", got)
		}
		if got := buff.Get(2); got != 0x04 {
			t.Fatalf("expected byte 0x04, got 0x%02x", got)
		}
	})

	t.Run("Add multiple bytes to a full ring buffer", func(t *testing.T) {
		buff := NewRingBuffer(3)
		buff.AddAll(0x01, 0x02, 0x03)
		buff.AddAll(0x04, 0x05, 0x06)

		if got := buff.Get(0); got != 0x04 {
			t.Fatalf("expected byte 0x04, got 0x%02x", got)
		}
		if got := buff.Get(1); got != 0x05 {
			t.Fatalf("expected byte 0x05, got 0x%02x", got)
		}
		if got := buff.Get(2); got != 0x06 {
			t.Fatalf("expected byte 0x06, got 0x%02x", got)
		}
	})
}
