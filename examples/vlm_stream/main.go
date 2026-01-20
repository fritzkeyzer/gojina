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
		Stream: true,
	}

	fmt.Println("Streaming response...")
	err := client.VLMStream(context.Background(), req, func(resp *jina.VLMResponse) error {
		for _, choice := range resp.Choices {
			fmt.Print(choice.Message.Content.Text)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("VLMStream error: %v", err)
	}
	fmt.Println("\nStream finished.")
}
