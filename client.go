package jina

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type config struct {
	APIKey       string
	EUCompliance bool
}

func defaultConfig() *config {
	return &config{
		APIKey:       "",
		EUCompliance: false,
	}
}

type Option func(*config)

type Client struct {
	cfg *config
}

func NewClient(options ...Option) *Client {
	cfg := defaultConfig()
	for _, option := range options {
		option(cfg)
	}

	return &Client{
		cfg: cfg,
	}
}

func WithAPIKey(apiKey string) Option {
	return func(cfg *config) {
		cfg.APIKey = apiKey
	}
}

func WithEUCompliance() Option {
	return func(cfg *config) {
		cfg.EUCompliance = true
	}
}

// doStream executes a streaming request and calls the callback for each data chunk.
func (cl *Client) doStream(req *http.Request, callback func([]byte) error) error {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error with status code: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return nil
			}
			if err := callback([]byte(data)); err != nil {
				return err
			}
		}
	}
	if errors.Is(scanner.Err(), io.EOF) {
		return nil
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
