package main

import (
	"encoding/binary"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	pool := NewRingBuffer(4096)

	// Initialize a Gin router with default Logger and Recovery middleware
	router := gin.Default()

	// GET random from the entropy pool
	router.GET("/random", func(c *gin.Context) {

		rnBytes, err := pool.Read(4)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		rnBin := binary.BigEndian.Uint32(rnBytes)
		rnHex := hex.EncodeToString(rnBytes)
		rmcount, _ := pool.Stats()

		rn := RandomNumber{
			Number:        rnBin,
			HexValue:      rnHex,
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

	// // Basic GET endpoint
	// router.GET("/ping", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"message": "pong",
	// 	})
	// })

	// // GET endpoint with a URL path parameter
	// router.GET("/hello/:name", func(c *gin.Context) {
	// 	name := c.Param("name")
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"greeting": "Hello, " + name + "!",
	// 	})
	// })

	// // POST endpoint that binds and validates incoming JSON
	// router.POST("/items", func(c *gin.Context) {
	// 	var newItem Item

	// 	if err := c.ShouldBindJSON(&newItem); err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 		return
	// 	}

	// 	c.JSON(http.StatusCreated, newItem)
	// })

	// Start the server on port 8080 (default)
	router.Run(":8080")
}
