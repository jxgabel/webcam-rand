package main

import (
	"encoding/binary"
	"encoding/hex"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const DefaultMin = 0
const MaxAllowedLimit = 1000000

func main() {

	// Create the entropy pool
	pool := NewRingBuffer(4096)

	// Initialize a Gin router with default Logger and Recovery middleware
	router := gin.Default()

	// GET random from the entropy pool
	router.GET("/random", func(c *gin.Context) {

		minStr := c.DefaultQuery("min", strconv.Itoa(DefaultMin))
		min, err := strconv.Atoi(minStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if min < DefaultMin {
			c.JSON(http.StatusBadRequest, gin.H{"error": "min must be 0 or greater"})
			return
		}

		maxStr := c.DefaultQuery("max", strconv.Itoa(MaxAllowedLimit))
		max, err := strconv.Atoi(maxStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if max > MaxAllowedLimit {
			c.JSON(http.StatusBadRequest, gin.H{"error": "max must be 1000000 or less"})
			return
		}

		if min >= max {
			c.JSON(http.StatusBadRequest, gin.H{"error": "min must be lower than max"})
			return
		}

		// Fetch data from entropy pool
		rnum, rbytes, err := GetBytes(pool, min, max)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		rnHex := hex.EncodeToString(rbytes)
		rmcount, _ := pool.Stats()

		rn := RandomNumber{
			Number:        rnum,
			HexValue:      rnHex,
			Min:           min,
			Max:           max,
			PoolRemaining: rmcount,
		}

		c.JSON(http.StatusOK, rn)
	})

	// GET the entropy pool status
	router.GET("/status", func(c *gin.Context) {

		pcnt, pcap := pool.Stats()
		pf := (float64(pcnt) / float64(pcap)) * 100

		PStatus := PoolStatus{
			AvailableBytes: pcnt,
			CapacityBytes:  pcap,
			PercentageFull: pf,
		}

		c.JSON(http.StatusOK, PStatus)
	})

	// POST incoming hashes from harvester service
	router.POST("/hash", func(c *gin.Context) {
		var newImageHash ImageHash

		if err := c.ShouldBindJSON(&newImageHash); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		dHash, err := hex.DecodeString(newImageHash.Hash)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(dHash) != 32 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hash length: expected 32 bytes"})
			return
		}

		pool.Write(dHash)

		c.JSON(http.StatusCreated, gin.H{"bytes_added": len(dHash)})
	})

	// Start the server on port 8080 (default)
	router.Run(":8888")
}

func GetBytes(pool *RingBuffer, min, max int) (uint32, []byte, error) {

	span := uint32(max - min + 1)
	limit := math.MaxUint32 - (math.MaxUint32 % span)

	for {

		rawBytes, err := pool.Read(4)
		if err != nil {
			return 0, nil, err
		}

		rawBin := binary.BigEndian.Uint32(rawBytes)

		if rawBin < limit {
			return uint32(min) + (rawBin % span), rawBytes, nil
		}
	}
}
