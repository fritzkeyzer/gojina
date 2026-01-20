package jina

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const DeepSearchModelDefault = "jina-deepsearch-v1"

type DeepSearchRequest struct {
	// Model is the ID of the model to use. Default: jina-deepsearch-v1.
	Model string `json:"model"`

	// Messages is the list of messages comprising the conversation.
	Messages []VLMMessage `json:"messages"`

	// Stream delivers events as they occur. Recommended for DeepSearch.
	Stream bool `json:"stream,omitempty"`

	// ReasoningEffort constrains effort on reasoning. Options: low, medium, high.
	ReasoningEffort string `json:"reasoning_effort,omitempty"`

	// BudgetTokens determines the maximum number of tokens allowed for the DeepSearch process.
	BudgetTokens int `json:"budget_tokens,omitempty"`

	// MaxAttempts is the maximum number of retries for solving a problem.
	MaxAttempts int `json:"max_attempts,omitempty"`

	// NoDirectAnswer forces the model to take further thinking/search steps even for trivial queries.
	NoDirectAnswer bool `json:"no_direct_answer,omitempty"`

	// MaxReturnedURLs is the maximum number of URLs to include in the final answer.
	MaxReturnedURLs int `json:"max_returned_urls,omitempty"`

	// ResponseFormat specifies the structured output format (JSON Schema).
	ResponseFormat *DeepSearchResponseFormat `json:"response_format,omitempty"`

	// BoostHostnames boosts specific hostnames in the search results.
	BoostHostnames []string `json:"boost_hostnames,omitempty"`
}

type DeepSearchResponseFormat struct {
	Type       string          `json:"type"` // e.g., "json_schema"
	JSONSchema json.RawMessage `json:"json_schema"`
}

// DeepSearchResponse represents the response from the DeepSearch API.
// Note: DeepSearch often streams, but this struct supports non-streaming or full accumulation.
type DeepSearchResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []DeepSearchChoice `json:"choices"`
}

type DeepSearchChoice struct {
	Index int `json:"index"`
	Delta struct {
		Content string `json:"content"`
		Type    string `json:"type"`
	} `json:"delta"`
	Message      VLMMessage `json:"message"`
	Logprobs     any        `json:"logprobs"`
	FinishReason string     `json:"finish_reason"`
}

// DeepSearch calls the Jina DeepSearch API for comprehensive investigation.
func (cl *Client) DeepSearch(ctx context.Context, req DeepSearchRequest) (*DeepSearchResponse, error) {
	url := "https://deepsearch.jina.ai/v1/chat/completions"

	if req.Model == "" {
		req.Model = DeepSearchModelDefault
	}
	req.Stream = false // Force stream to false for synchronous call

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

	var result DeepSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeepSearchStream calls the Jina DeepSearch API with streaming enabled.
// The callback function is invoked for each chunk of the response.
func (cl *Client) DeepSearchStream(ctx context.Context, req DeepSearchRequest, callback func(*DeepSearchResponse) error) error {
	url := "https://deepsearch.jina.ai/v1/chat/completions"

	if req.Model == "" {
		req.Model = DeepSearchModelDefault
	}
	req.Stream = true

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if cl.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+cl.cfg.APIKey)
	}

	return cl.doStream(httpReq, func(data []byte) error {
		//fmt.Println("data: ", string(data))
		var chunk DeepSearchResponse
		if err := json.Unmarshal(data, &chunk); err != nil {
			return fmt.Errorf("failed to unmarshal chunk: %w", err)
		}
		return callback(&chunk)
	})
}
