package main

import (
	"sync"
)

type ImageHash struct {
	Hash string `json:"hash" binding:"required"`
}

type RandomNumber struct {
	Number        uint32 `json:"number"`
	HexValue      string `json:"hex"`
	Min           int    `json:"min"`
	Max           int    `json:"max"`
	PoolRemaining int    `json:"pool_remaining"`
}

type PoolStatus struct {
	AvailableBytes int     `json:"available_bytes"`
	CapacityBytes  int     `json:"capacity_bytes"`
	PercentageFull float64 `json:"percentage_full"`
}

type RingBuffer struct {
	Mu       sync.Mutex
	Data     []byte
	Capacity int
	WriteIdx int
	ReadIdx  int
	Count    int
}
