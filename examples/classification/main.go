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

	req := jina.ClassificationRequest{
		Model: jina.ClassificationModelEmbeddingsV3,
		Input: []jina.ClassificationInput{
			jina.NewClassificationInputText("The product is amazing!"),
			jina.NewClassificationInputText("I am very disappointed."),
		},
		Labels: []string{"Positive", "Negative"},
	}

	resp, err := client.Classify(context.Background(), req)
	if err != nil {
		log.Fatalf("Classification error: %v", err)
	}

	for _, data := range resp.Data {
		fmt.Printf("Input Index: %d, Prediction: %s, Score: %f\n", data.Index, data.Prediction, data.Score)
	}
}
