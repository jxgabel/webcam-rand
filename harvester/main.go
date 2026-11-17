package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"webcam-rand/harvester/config"
)

type HashPayload struct {
	Hash string `json:"hash"`
}

func main() {

	// Load Config
	cfg, err := config.Load("config/config.json")
	if err != nil {
		log.Fatalf("Startup error: %v", err)
	}

	// Create Context for background camera polling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Open cameras
	cams, err := OpenCameras(ctx, cfg.Sources)
	if len(cams) == 0 {
		log.Printf("error: %v", err)
		log.Fatal(err)
	}

	// Create a context that cancels automatically on Ctrl+C (SIGINT) or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Setup ticker
	postHashURL := cfg.Api + "/hash"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	ticker := time.NewTicker(time.Duration(cfg.CaptureIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	fmt.Println("Service started. Press Ctrl+C to stop.")

	// Race the ticker channel against the context's Done channel
	for {
		select {
		case <-ticker.C:

			// Put your repeating harvester work here
			fmt.Println("Doing work...")

			frames, err := CaptureFrames(cams)
			if err != nil {
				log.Printf("error: %v", err)
				continue
			}

			hexHash, err := ProcessSourceData(frames)
			if err != nil {
				log.Printf("error: %v", err)
				continue
			}

			log.Printf("hash: %v", hexHash)

			hp := HashPayload{
				Hash: hexHash,
			}

			hPL, err := json.Marshal(hp)
			if err != nil {
				log.Printf("error: %v", err)
				continue
			}

			rdr := bytes.NewBuffer(hPL)

			req, err := http.NewRequest("POST", postHashURL, rdr)
			if err != nil {
				log.Printf("error: %v", err)
				continue
			}

			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)

			if err != nil {
				log.Printf("error: %v", err)
				continue
			}

			respBody, err := io.ReadAll(resp.Body)
			log.Printf("POST Response Body: %s", respBody)

			if resp.StatusCode != http.StatusCreated {
				log.Printf("Data was not submitted. (Status Code: %v)", resp.StatusCode)
			}

			resp.Body.Close()

		case <-ctx.Done():
			// Triggered when the interrupt signal arrives
			fmt.Println("Shutdown signal received, exiting loop.")
			return
		}
	}

}
