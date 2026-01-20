package jina

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmbeddingModel string

const (
	// EmbeddingModelV4 is a 3.8B parameter multimodal and multilingual embedding model.
	// Supports text, images, and PDFs. Output dimensions: 2048.
	EmbeddingModelV4 EmbeddingModel = "jina-embeddings-v4"
	// EmbeddingModelV3 is a 570M parameter multilingual text embedding model.
	// Output dimensions: 1024.
	EmbeddingModelV3 EmbeddingModel = "jina-embeddings-v3"
	// EmbeddingModelClipV2 is a 885M parameter multimodal embedding model.
	// Best for cross-modal text-image retrieval. Output dimensions: 1024.
	EmbeddingModelClipV2 EmbeddingModel = "jina-clip-v2"
	// EmbeddingModelCode0_5B is a 494M code embedding model.
	EmbeddingModelCode0_5B EmbeddingModel = "jina-code-embeddings-0.5b"
	// EmbeddingModelCode1_5B is a 1.54B code embedding model.
	EmbeddingModelCode1_5B EmbeddingModel = "jina-code-embeddings-1.5b"
)

type EmbeddingTask string

const (
	EmbeddingTaskRetrievalQuery         EmbeddingTask = "retrieval.query"
	EmbeddingTaskRetrievalPassage       EmbeddingTask = "retrieval.passage"
	EmbeddingTaskTextMatching           EmbeddingTask = "text-matching"
	EmbeddingTaskCodeQuery              EmbeddingTask = "code.query"
	EmbeddingTaskCodePassage            EmbeddingTask = "code.passage"
	EmbeddingTaskClassification         EmbeddingTask = "classification"
	EmbeddingTaskSeparation             EmbeddingTask = "separation"
	EmbeddingTaskNL2CodeQuery           EmbeddingTask = "nl2code.query"
	EmbeddingTaskNL2CodePassage         EmbeddingTask = "nl2code.passage"
	EmbeddingTaskCode2CodeQuery         EmbeddingTask = "code2code.query"
	EmbeddingTaskCode2CodePassage       EmbeddingTask = "code2code.passage"
	EmbeddingTaskCode2NLQuery           EmbeddingTask = "code2nl.query"
	EmbeddingTaskCode2NLPassage         EmbeddingTask = "code2nl.passage"
	EmbeddingTaskCode2CompletionQuery   EmbeddingTask = "code2completion.query"
	EmbeddingTaskCode2CompletionPassage EmbeddingTask = "code2completion.passage"
	EmbeddingTaskQAQuery                EmbeddingTask = "qa.query"
	EmbeddingTaskQAPassage              EmbeddingTask = "qa.passage"
)

type EmbeddingsRequest struct {
	// Model is the identifier of the model to use.
	Model EmbeddingModel `json:"model"`

	// Input is the array of input strings or objects to be embedded.
	// Use NewEmbeddingInputText, NewEmbeddingInputImage, or NewEmbeddingInputPDF.
	Input []EmbeddingInput `json:"input"`

	// EmbeddingType specifies the format of the returned embeddings.
	// Options: float, base64, binary, ubinary. Default: float.
	EmbeddingType []string `json:"embedding_type,omitempty"`

	// Task specifies the intended downstream application.
	Task EmbeddingTask `json:"task,omitempty"`

	// Dimensions truncates output embeddings to the specified size if set.
	Dimensions int `json:"dimensions,omitempty"`

	// LateChunking, if true, concatenates all sentences in input and treats as a single input.
	LateChunking bool `json:"late_chunking,omitempty"`

	// Truncate, if true, automatically drops the tail that extends beyond max context length.
	Truncate bool `json:"truncate,omitempty"`

	// ReturnMultivector, if true, returns NxD multi-vector embeddings (Only for v4).
	ReturnMultivector bool `json:"return_multivector,omitempty"`

	// Normalized, if true, embeddings are normalized to unit L2 norm (Only for v3/clip).
	Normalized bool `json:"normalized,omitempty"`
}

// EmbeddingInput represents a single input item for embeddings.
// It uses custom marshaling to send simple strings or JSON objects as required.
type EmbeddingInput struct {
	Text  string `json:"text,omitempty"`
	Image string `json:"image,omitempty"`
	PDF   string `json:"pdf,omitempty"`
}

// MarshalJSON implements custom marshaling to support both string and object formats.
func (e EmbeddingInput) MarshalJSON() ([]byte, error) {
	if e.Image != "" {
		return json.Marshal(map[string]string{"image": e.Image})
	}
	if e.PDF != "" {
		return json.Marshal(map[string]string{"pdf": e.PDF})
	}
	// Default to simple string for text to support code models and simplicity
	return json.Marshal(e.Text)
}

func NewEmbeddingInputText(text string) EmbeddingInput {
	return EmbeddingInput{Text: text}
}

func NewEmbeddingInputImage(imageURLOrBase64 string) EmbeddingInput {
	return EmbeddingInput{Image: imageURLOrBase64}
}

// NewEmbeddingInputPDF creates a PDF input for embeddings (v4 only).
func NewEmbeddingInputPDF(pdfURL string) EmbeddingInput {
	return EmbeddingInput{PDF: pdfURL}
}

type EmbeddingsResponse struct {
	Data  []EmbeddingData `json:"data"`
	Usage Usage           `json:"usage"`
}

type EmbeddingData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

type Usage struct {
	TotalTokens      int `json:"total_tokens"`
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
}

// Embeddings calls the Jina Embeddings API.
func (cl *Client) Embeddings(ctx context.Context, req EmbeddingsRequest) (*EmbeddingsResponse, error) {
	url := "https://api.jina.ai/v1/embeddings"

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

	resp, err := cl.do(httpReq)
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

	var result EmbeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Helper method to execute requests (can be moved to client.go later)
func (cl *Client) do(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	return client.Do(req)
}
