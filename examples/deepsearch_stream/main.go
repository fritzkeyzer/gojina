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

	req := jina.DeepSearchRequest{
		Model: "jina-deepsearch-v1",
		Messages: []jina.VLMMessage{
			jina.NewVLMMessage("user", "what is the latest blog post from jina ai?"),
		},
		Stream: true,
	}

	fmt.Println("Streaming response...")
	err := client.DeepSearchStream(context.Background(), req, func(resp *jina.DeepSearchResponse) error {
		fmt.Print(resp.Choices[0].Delta.Content)
		return nil
	})

	if err != nil {
		log.Fatalf("DeepSearchStream error: %v", err)
	}
	fmt.Println("\nStream finished.")
}
