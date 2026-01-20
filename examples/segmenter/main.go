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

	req := jina.SegmenterRequest{
		Content:      "Jina AI is the search foundation for the age of generative AI. It provides APIs for embeddings, reranker, reader, and search.",
		Tokenizer:    "cl100k_base",
		ReturnChunks: true,
	}

	resp, err := client.Segment(context.Background(), req)
	if err != nil {
		log.Fatalf("Segmenter error: %v", err)
	}

	fmt.Printf("Number of tokens: %d\n", resp.NumTokens)
	fmt.Printf("Number of chunks: %d\n", resp.NumChunks)
	for i, chunk := range resp.Chunks {
		fmt.Printf("Chunk %d: %s\n", i, chunk)
	}
}
