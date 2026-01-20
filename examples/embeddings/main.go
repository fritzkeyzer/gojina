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

	req := jina.EmbeddingsRequest{
		Model: jina.EmbeddingModelV3,
		Input: []jina.EmbeddingInput{
			jina.NewEmbeddingInputText("Hello, world!"),
			jina.NewEmbeddingInputText("Jina AI is awesome"),
		},
		Task: jina.EmbeddingTaskRetrievalPassage,
	}

	resp, err := client.Embeddings(context.Background(), req)
	if err != nil {
		log.Fatalf("Embeddings error: %v", err)
	}

	for _, data := range resp.Data {
		fmt.Printf("Index: %d, Embedding length: %d\n", data.Index, len(data.Embedding))
	}
	fmt.Printf("Total tokens: %d\n", resp.Usage.TotalTokens)
}
