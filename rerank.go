package jina

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type RerankerModel string

const (
	// RerankerModelV3 is a 0.6B parameter multilingual document reranker.
	// Uses causal self-attention for better context.
	RerankerModelV3 RerankerModel = "jina-reranker-v3"
	// RerankerModelM0 is a 2.4B parameter multimodal reranker (text and images).
	RerankerModelM0 RerankerModel = "jina-reranker-m0"
	// RerankerModelV2BaseMultilingual is a 278M parameter multilingual text reranker.
	RerankerModelV2BaseMultilingual RerankerModel = "jina-reranker-v2-base-multilingual"
	// RerankerModelColbertV2 is a 560M parameter ColBERT-style reranker.
	RerankerModelColbertV2 RerankerModel = "jina-colbert-v2"
)

// RerankInput represents a document or query object containing text and/or image.
// Used for multimodal models (like M0) or when structured input is preferred.
type RerankInput struct {
	Text  string `json:"text,omitempty"`
	Image string `json:"image,omitempty"`
}

// RerankRequest is the request body for the Rerank API.
// It supports both simple text/string inputs and structured multimodal inputs via separate fields.
// The MarshalJSON method ensures the correct JSON structure is sent to the API.
type RerankRequest struct {
	// Model is the identifier of the model to use.
	Model RerankerModel `json:"model"`

	// Simple usage (Text or Image URL as string)

	// Query is the search query string.
	// Use this for text queries (v3, v2) or image URLs (m0).
	Query string `json:"-"`

	// Documents is the list of text strings to rerank.
	// Use this for standard text reranking (v3, v2).
	Documents []string `json:"-"`

	// Structured/Multimodal usage

	// QueryInput allows passing a structured query object (e.g., {text: "..."} or {image: "..."}).
	// Use this if you need to pass specific object fields for the query.
	// If non-nil, this takes precedence over Query string.
	QueryInput *RerankInput `json:"-"`

	// DocumentsInput is the list of structured documents to rerank.
	// Use this for multimodal documents (m0) or when you need to pass objects.
	// If non-empty, this takes precedence over Documents string list.
	DocumentsInput []RerankInput `json:"-"`

	// TopN is the number of most relevant documents to return.
	// Defaults to the length of documents.
	TopN int `json:"top_n,omitempty"`

	// ReturnDocuments decides whether to return the document text/content.
	// Default is true. Use pointer to distinguish omitted vs false.
	ReturnDocuments *bool `json:"return_documents,omitempty"`
}

// MarshalJSON implements custom marshaling to map the distinct Go fields to the unified JSON API structure.
func (r RerankRequest) MarshalJSON() ([]byte, error) {
	// 1. Create a map to hold the JSON data.
	data := make(map[string]interface{})

	data["model"] = r.Model
	if r.TopN > 0 {
		data["top_n"] = r.TopN
	}
	if r.ReturnDocuments != nil {
		data["return_documents"] = r.ReturnDocuments
	}

	// 2. Handle Query
	if r.QueryInput != nil {
		data["query"] = r.QueryInput
	} else {
		data["query"] = r.Query
	}

	// 3. Handle Documents
	if len(r.DocumentsInput) > 0 {
		data["documents"] = r.DocumentsInput
	} else {
		data["documents"] = r.Documents
	}

	return json.Marshal(data)
}

type RerankResponse struct {
	Model   string         `json:"model"`
	Usage   Usage          `json:"usage"`
	Results []RerankResult `json:"results"`
}

type RerankResult struct {
	Index          int             `json:"index"`
	RelevanceScore float64         `json:"relevance_score"`
	Document       json.RawMessage `json:"document,omitempty"` // Returns the input document (string or object)
}

// Rerank calls the Jina Reranker API to rank documents based on relevance to the query.
func (cl *Client) Rerank(ctx context.Context, req RerankRequest) (*RerankResponse, error) {
	url := "https://api.jina.ai/v1/rerank"

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if cl.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+cl.cfg.APIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("API error: %v", errResp)
		}
		return nil, fmt.Errorf("API error with status code: %d", resp.StatusCode)
	}

	var result RerankResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
