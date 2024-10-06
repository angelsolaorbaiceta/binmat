package signature

import "io"

// A RingBuffer is a fixed-size buffer that discards the oldest bytes when
// new bytes are written to it and it is full.
//
// Bytes that aren't set yet are returned as 0.
type RingBuffer struct {
	data []byte
	// capacity is the maximum number of bytes that the buffer can hold.
	capacity int
	// start is the index of the oldest byte in the buffer.
	start int
	// size is the number of bytes currently in the buffer.
	size int
}

// A ByteBuffer is a byte buffer that can be read from and written to.
type ByteBuffer interface {
	Size() int
	Capacity() int
	Get(idx int) byte
	Add(b byte)
}

func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}

	return &RingBuffer{
		data:     make([]byte, capacity),
		capacity: capacity,
		start:    0,
		size:     0,
	}
}

func (buff *RingBuffer) Size() int {
	return buff.size
}

func (buff *RingBuffer) Capacity() int {
	return buff.capacity
}

// Data returns a copy of the data in the buffer, in the correct order.
func (buff *RingBuffer) Data() []byte {
	data := make([]byte, buff.size)
	for i := 0; i < buff.size; i++ {
		data[i] = buff.Get(i)
	}

	return data
}

// Get returns the byte at the given index in the buffer.
//
// If the index is greater than or equal to the size of the buffer, 0 is returned.
// That is, non-set bytes are returned as 0.
//
// If the index is greater than or equal to the capacity of the buffer, the function panics.
// That is considered a logic error that's non-recoverable.
func (buff *RingBuffer) Get(idx int) byte {
	if idx >= buff.capacity {
		panic("index out of bounds")
	}

	if idx >= buff.size {
		return 0
	}

	return buff.data[(buff.start+idx)%buff.capacity]
}

// Add adds a single byte to the buffer.
// When the buffer is already full, the oldest byte is discarded.
func (buff *RingBuffer) Add(b byte) {
	if buff.size < buff.capacity {
		// Buffer not full: simply add the byte
		buff.data[(buff.start+buff.size)%buff.capacity] = b
		buff.size++
	} else {
		// Buffer full: overwrite the oldest byte and move the start index
		buff.data[buff.start] = b
		buff.start = (buff.start + 1) % buff.capacity
	}
}

// AddAll adds multiple bytes to the buffer.
func (buff *RingBuffer) AddAll(bytes ...byte) {
	for _, b := range bytes {
		buff.Add(b)
	}
}

// AddFrom adds multiple bytes from a byte slice to the buffer.
func (buff *RingBuffer) AddFrom(bytes []byte) {
	for _, b := range bytes {
		buff.Add(b)
	}
}

// Read reads at most n bytes from the given reader and adds them to the buffer.
func (buff *RingBuffer) Read(r io.Reader, n int) (int, error) {
	temp := make([]byte, n)
	m, err := r.Read(temp)
	if err != nil {
		return m, err
	}

	buff.AddFrom(temp)

	return m, nil
}
