package main

import (
	"errors"
)

func NewRingBuffer(capacity int) *RingBuffer {

	rb := RingBuffer{
		Data:     make([]byte, capacity),
		Capacity: capacity,
	}

	return &rb
}

func (rb *RingBuffer) Read(n int) ([]byte, error) {

	rb.Mu.Lock()
	defer rb.Mu.Unlock()

	if rb.Count < n {
		return nil, errors.New("Not enough bytes to read")
	}

	rn := make([]byte, n)

	for i := range n {

		rn[i] = rb.Data[rb.ReadIdx]
		rb.ReadIdx = (rb.ReadIdx + 1) % rb.Capacity
		rb.Count--
	}

	return rn, nil
}

func (rb *RingBuffer) Write(incoming []byte) {

	rb.Mu.Lock()
	defer rb.Mu.Unlock()

	for _, b := range incoming {

		rb.Data[rb.WriteIdx] = b
		rb.WriteIdx = (rb.WriteIdx + 1) % rb.Capacity

		if rb.Count == rb.Capacity {
			rb.ReadIdx = (rb.ReadIdx + 1) % rb.Capacity
		} else {
			rb.Count++
		}
	}
}

func (rb *RingBuffer) Stats() (int, int) {

	rb.Mu.Lock()
	defer rb.Mu.Unlock()

	return rb.Count, rb.Capacity
}
