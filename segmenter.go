package jina

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type SegmenterRequest struct {
	// Content is the text content to segment.
	Content string `json:"content"`

	// Tokenizer specifies the tokenizer to use.
	// Options: cl100k_base, o200k_base, p50k_base, r50k_base, p50k_edit, gpt2.
	// Default: cl100k_base.
	Tokenizer string `json:"tokenizer,omitempty"`

	// ReturnTokens, if true, includes tokens and their IDs in the response.
	ReturnTokens bool `json:"return_tokens,omitempty"`

	// ReturnChunks, if true, segments the text into semantic chunks.
	ReturnChunks bool `json:"return_chunks,omitempty"`

	// MaxChunkLength is the maximum characters per chunk (only effective if ReturnChunks is true).
	// Default: 1000.
	MaxChunkLength int `json:"max_chunk_length,omitempty"`

	// Head returns the first N tokens (exclusive with Tail).
	Head int `json:"head,omitempty"`

	// Tail returns the last N tokens (exclusive with Head).
	Tail int `json:"tail,omitempty"`
}

type SegmenterResponse struct {
	NumTokens      int       `json:"num_tokens"`
	Tokenizer      string    `json:"tokenizer"`
	Usage          Usage     `json:"usage"`
	NumChunks      int       `json:"num_chunks,omitempty"`
	ChunkPositions [][]int   `json:"chunk_positions,omitempty"`
	Tokens         [][]Token `json:"tokens,omitempty"` // List of chunks, each containing a list of Tokens
	Chunks         []string  `json:"chunks,omitempty"`
}

// Token represents a single token with its text and ID(s).
// It corresponds to the format: ["token_text", [id1, id2...]]
type Token struct {
	Text string
	IDs  []int
}

// UnmarshalJSON implements custom unmarshaling for Token.
func (t *Token) UnmarshalJSON(data []byte) error {
	var raw []interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) != 2 {
		return fmt.Errorf("invalid token format: expected 2 elements, got %d", len(raw))
	}

	text, ok := raw[0].(string)
	if !ok {
		return fmt.Errorf("invalid token format: first element is not string")
	}
	t.Text = text

	// Second element is array of IDs
	idsRaw, ok := raw[1].([]interface{})
	if !ok {
		return fmt.Errorf("invalid token format: second element is not array")
	}

	t.IDs = make([]int, len(idsRaw))
	for i, idRaw := range idsRaw {
		idFloat, ok := idRaw.(float64)
		if !ok {
			return fmt.Errorf("invalid token ID: not a number")
		}
		t.IDs[i] = int(idFloat)
	}
	return nil
}

// Segment calls the Jina Segmenter API to tokenize or chunk text.
func (cl *Client) Segment(ctx context.Context, req SegmenterRequest) (*SegmenterResponse, error) {
	url := "https://segment.jina.ai/"

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

	var result SegmenterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
