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

	req := jina.VLMRequest{
		Model: "jina-vlm",
		Messages: []jina.VLMMessage{
			jina.NewVLMMessageWithParts("user", []jina.VLMContentPart{
				{
					Type: "text",
					Text: "Describe this image",
				},
				{
					Type:     "image_url",
					ImageURL: &jina.VLMImageURL{URL: "https://jina-ai-gmbh.ghost.io/content/images/size/w1600/2025/10/image-1.png"},
				},
			}),
		},
	}

	resp, err := client.VLM(context.Background(), req)
	if err != nil {
		log.Fatalf("VLM error: %v", err)
	}

	for _, choice := range resp.Choices {
		fmt.Printf("Response: %s\n", choice.Message.Content.Text)
	}
}
