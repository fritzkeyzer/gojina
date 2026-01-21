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

	req := jina.SearchRequest{
		Query:        "Jina AI",
		JSONResponse: false,
		MaxResults:   10,
		CountryCode:  "DE",
		LanguageCode: "de",
	}

	resp, err := client.Search(context.Background(), req)
	if err != nil {
		log.Fatalf("Search error: %v", err)
	}

	if resp.Structured != nil {
		for i, item := range resp.Structured.Data {
			fmt.Printf("Result %d: %s (%s)\n", i+1, item.Title, item.URL)
		}
	} else {
		fmt.Println(resp.Text)
	}
}
