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

	// Example 1: Basic text reranking (using simple string fields)
	reqV3 := jina.RerankRequest{
		Model: jina.RerankerModelV3,
		Query: "programming languages",
		Documents: []string{
			"Python is a high-level general-purpose programming language.",
			"Jina AI offers state-of-the-art embeddings.",
			"The quick brown fox jumps over the lazy dog.",
		},
		TopN: 2,
	}

	fmt.Println("--- Text Reranking (V3) ---")
	respV3, err := client.Rerank(context.Background(), reqV3)
	if err != nil {
		log.Fatalf("Rerank V3 error: %v", err)
	}
	for _, result := range respV3.Results {
		fmt.Printf("Index: %d, Score: %f, Document: %s\n", result.Index, result.RelevanceScore, result.Document)
	}
	fmt.Println()

	// Example 2: Multimodal reranking (using structured input fields)
	// Requires jina-reranker-m0
	reqM0 := jina.RerankRequest{
		Model: jina.RerankerModelM0,
		// Query can be text or image string
		Query: "https://jina-ai-gmbh.ghost.io/content/images/2025/10/Heading--54-.svg", // jina reranker 3 model architecture diagram
		// Documents can be mixed text/image objects
		DocumentsInput: []jina.RerankInput{
			{Text: "Python is a high-level general-purpose programming language."},
			{Image: "https://jina-ai-gmbh.ghost.io/content/images/size/w1600/2025/10/image-1.png"}, // benchmark results image
			{Text: "The quick brown fox jumps over the lazy dog."},
			{Text: "Jina AI offers state-of-the-art embeddings."},
			{Text: "Jina AI offers state-of-the-art re-ranking models"},
		},
		TopN: 3,
	}

	fmt.Println("--- Multimodal Reranking (M0) ---")
	respM0, err := client.Rerank(context.Background(), reqM0)
	if err != nil {
		log.Fatalf("Rerank M0 error: %v", err)
	}
	for _, result := range respM0.Results {
		fmt.Printf("Index: %d, Score: %f, Document: %s\n", result.Index, result.RelevanceScore, result.Document)
	}
}
