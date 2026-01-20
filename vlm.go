package jina

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const VLMModelDefault = "jina-vlm"

type VLMRequest struct {
	// Model is the identifier of the model to use. Default: jina-vlm.
	Model string `json:"model"`

	// Messages is a list of messages comprising the conversation.
	Messages []VLMMessage `json:"messages"`

	// Stream, if true, returns tokens as they are generated via server-sent events.
	Stream bool `json:"stream,omitempty"`
}

type VLMMessage struct {
	Role    string            `json:"role"`
	Content VLMMessageContent `json:"content"`
}

// VLMMessageContent represents the content of a message, which can be a simple string or a list of parts.
type VLMMessageContent struct {
	Text  string           `json:"-"`
	Parts []VLMContentPart `json:"-"`
}

// MarshalJSON implements custom marshaling for VLMMessageContent.
func (c VLMMessageContent) MarshalJSON() ([]byte, error) {
	if len(c.Parts) > 0 {
		return json.Marshal(c.Parts)
	}
	return json.Marshal(c.Text)
}

// UnmarshalJSON implements custom unmarshaling for VLMMessageContent.
func (c *VLMMessageContent) UnmarshalJSON(data []byte) error {
	// Try unmarshaling as string first
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		c.Text = text
		return nil
	}

	// Try unmarshaling as parts
	var parts []VLMContentPart
	if err := json.Unmarshal(data, &parts); err == nil {
		c.Parts = parts
		return nil
	}

	return fmt.Errorf("invalid VLMMessageContent: not a string or array of parts")
}

func NewVLMMessage(role, text string) VLMMessage {
	return VLMMessage{
		Role: role,
		Content: VLMMessageContent{
			Text: text,
		},
	}
}

func NewVLMMessageWithParts(role string, parts []VLMContentPart) VLMMessage {
	return VLMMessage{
		Role: role,
		Content: VLMMessageContent{
			Parts: parts,
		},
	}
}

// VLMContentPart represents a part of the message content (text or image).
type VLMContentPart struct {
	Type     string       `json:"type"`
	Text     string       `json:"text,omitempty"`
	ImageURL *VLMImageURL `json:"image_url,omitempty"`
}

type VLMImageURL struct {
	URL string `json:"url"`
}

type VLMResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices []VLMChoice `json:"choices"`
	Usage   Usage       `json:"usage"`
}

type VLMChoice struct {
	Index        int        `json:"index"`
	Message      VLMMessage `json:"message"`
	FinishReason string     `json:"finish_reason"`
}

// VLM calls the Jina VLM API for image understanding and multimodal chat.
func (cl *Client) VLM(ctx context.Context, req VLMRequest) (*VLMResponse, error) {
	url := "https://api-beta-vlm.jina.ai/v1/chat/completions"

	if req.Model == "" {
		req.Model = VLMModelDefault
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

	var result VLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// VLMStream calls the Jina VLM API with streaming enabled.
// The callback function is invoked for each chunk of the response.
func (cl *Client) VLMStream(ctx context.Context, req VLMRequest, callback func(*VLMResponse) error) error {
	url := "https://api-beta-vlm.jina.ai/v1/chat/completions"

	if req.Model == "" {
		req.Model = VLMModelDefault
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
		var chunk VLMResponse
		if err := json.Unmarshal(data, &chunk); err != nil {
			return fmt.Errorf("failed to unmarshal chunk: %w", err)
		}
		return callback(&chunk)
	})
}
