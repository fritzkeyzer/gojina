package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fritzkeyzer/gojina"
)

func main() {
	// Get your Jina AI API key for free: https://jina.ai/?sui=apikey
	apiKey := os.Getenv("JINA_API_KEY")
	if apiKey == "" {
		log.Fatal("JINA_API_KEY environment variable is not set")
	}

	client := jina.NewClient(jina.WithAPIKey(apiKey))

	req := jina.ReaderRequest{
		URL:          "https://jina.ai",
		JSONResponse: true,
	}

	resp, err := client.Reader(context.Background(), req)
	if err != nil {
		log.Fatalf("Reader error: %v", err)
	}

	if resp.Structured != nil {
		fmt.Printf("Title: %s\n", resp.Structured.Data.Title)
		fmt.Printf("Content Preview: %s...\n", resp.Structured.Data.Content[:300])
	} else {
		fmt.Println("No structured data returned")
	}
}
