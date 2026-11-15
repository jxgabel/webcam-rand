package main

import (
	"crypto/sha256"
)

func ProcessSourceData (rawData [][]byte) (string, error) {

  if len(rawData) == 0 {
	return "", errors.New("no data to process")
  }

  h := sha256.New() // reusable hash object is more efficient
  hashBuffer:= make([]byte, 0, h.Size())  // temp storage for Sum
  xorBuffer := make([]byte, h.Size())

  //loop through the data slices
  for _, data := range rawData {
	h.Reset()
	h.Write(data)
    hash := h.Sum(hashBuffer)

	// on each iteration xor the data in to the buffer
    for i := range hash {
      xorBuffer[i] ^= hash[i]
	}

  }

  return hex.EncodeToString(xorBuffer), nil
}

