package jina

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ClassificationModel string

const (
	// ClassificationModelClipV2 is best for cross-modal text-image retrieval and zero-shot image classification.
	// Dimensions: 1024. Size: 885M.
	ClassificationModelClipV2 ClassificationModel = "jina-clip-v2"
	// ClassificationModelEmbeddingsV4 is a 3.8B parameter multimodal model.
	// Best for unified multimodal retrieval and document understanding. Dimensions: 2048.
	ClassificationModelEmbeddingsV4 ClassificationModel = "jina-embeddings-v4"
	// ClassificationModelEmbeddingsV3 is a 570M parameter multilingual text embedding model.
	// Optimized for retrieval, classification, and text matching. Dimensions: 1024.
	ClassificationModelEmbeddingsV3 ClassificationModel = "jina-embeddings-v3"
)

type ClassificationRequest struct {
	// Model is the identifier of the model to use.
	// Options: jina-clip-v2, jina-embeddings-v4, jina-embeddings-v3.
	// Required if ClassifierID is not provided.
	Model ClassificationModel `json:"model,omitempty"`

	// ClassifierID is the identifier of the classifier.
	// If not provided, a new classifier will be created.
	ClassifierID string `json:"classifier_id,omitempty"`

	// Input is the array of inputs for classification.
	// Use NewClassificationInputText or NewClassificationInputImage to create inputs.
	Input []ClassificationInput `json:"input"`

	// Labels is the list of labels used for classification.
	Labels []string `json:"labels"`
}

type ClassificationInput struct {
	Text  string `json:"text,omitempty"`
	Image string `json:"image,omitempty"`
}

// MarshalJSON implements custom marshaling to support both string and object formats.
func (c ClassificationInput) MarshalJSON() ([]byte, error) {
	if c.Image != "" {
		return json.Marshal(map[string]string{"image": c.Image})
	}
	// Default to simple string for text
	return json.Marshal(c.Text)
}

func NewClassificationInputText(text string) ClassificationInput {
	return ClassificationInput{Text: text}
}

// NewClassificationInputImage creates an image input for classification.
// imageURLOrBase64 can be a URL or a base64 encoded string.
func NewClassificationInputImage(imageURLOrBase64 string) ClassificationInput {
	return ClassificationInput{Image: imageURLOrBase64}
}

type ClassificationResponse struct {
	Data  []ClassificationData `json:"data"`
	Usage Usage                `json:"usage"`
}

type ClassificationData struct {
	Object      string                `json:"object"`
	Index       int                   `json:"index"`
	Prediction  string                `json:"prediction"`
	Score       float64               `json:"score"`
	Predictions []ClassificationLabel `json:"predictions,omitempty"`
}

type ClassificationLabel struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

// Classify calls the Jina Classifier API to classify text or images into categories.
func (cl *Client) Classify(ctx context.Context, req ClassificationRequest) (*ClassificationResponse, error) {
	url := "https://api.jina.ai/v1/classify"

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

	var result ClassificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
