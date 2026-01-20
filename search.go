package jina

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type SearchRequest struct {
	// Query is the search query.
	Query string `json:"q"`

	// CountryCode is the two-letter country code to use for the search (e.g., "us", "uk").
	CountryCode string `json:"gl,omitempty"`

	// Location is the location from where you want the search query to originate (e.g., city level).
	Location string `json:"location,omitempty"`

	// LanguageCode is the two-letter language code to use for the search (e.g., "en", "de").
	LanguageCode string `json:"hl,omitempty"`

	// MaxResults sets the maximum results returned.
	// Using this may cause latency and exclude specialized result types.
	MaxResults int `json:"num,omitempty"`

	// PageOffset is the result offset for pagination.
	PageOffset int `json:"page,omitempty"`

	// Header Options

	// JSONResponse controls whether to request application/json response.
	JSONResponse bool `json:"-"`

	// Site limits the search to a specific domain (e.g., "https://jina.ai").
	Site string `json:"-"`

	// WithLinksSummary, if true, gathers all links (or unique links) at the end of the response.
	WithLinksSummary bool `json:"-"`

	// WithImagesSummary, if true, gathers all images (or unique images) at the end of the response.
	WithImagesSummary bool `json:"-"`

	// RetainImages can be set to "none" to remove all images from the response.
	RetainImages string `json:"-"`

	// NoCache, if true, bypasses cache and retrieves real-time data.
	NoCache bool `json:"-"`

	// WithGeneratedAlt, if true, generates captions for images without alt tags.
	WithGeneratedAlt bool `json:"-"`

	// RespondWith can be "no-content" to exclude page content from the response.
	RespondWith string `json:"-"`

	// WithFavicon, if true, includes favicon of the website in the response.
	WithFavicon bool `json:"-"`

	// ReturnFormat specifies the return format: markdown, html, text, screenshot, pageshot.
	ReturnFormat string `json:"-"`

	// Engine specifies the engine: "browser" or "direct".
	Engine string `json:"-"`

	// WithFavicons, if true, fetches the favicon of each URL in the SERP.
	WithFavicons bool `json:"-"`

	// Timeout specifies the maximum time (in seconds) to wait for the webpage to load.
	Timeout int `json:"-"`

	// SetCookie forwards custom cookie settings when accessing the URL.
	SetCookie string `json:"-"`

	// ProxyURL utilizes your proxy to access URLs.
	ProxyURL string `json:"-"`

	// Locale controls the browser locale to render the page.
	Locale string `json:"-"`

	// EUCompliance, if true, uses EU infrastructure (eu.s.jina.ai).
	EUCompliance bool `json:"-"`
}

type SearchResponse struct {
	Text       string                    // Raw text response
	Structured *StructuredSearchResponse // Structured JSON response
}

type StructuredSearchResponse struct {
	Code   int                `json:"code"`
	Status int                `json:"status"`
	Data   []SearchResultData `json:"data"`
	Usage  struct {
		Tokens int `json:"tokens"`
	} `json:"usage"`
}

type SearchResultData struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Content     string `json:"content"`
	Usage       struct {
		Tokens int `json:"tokens"`
	} `json:"usage"`
}

// Search calls the Jina Search API to search the web.
func (cl *Client) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if cl.cfg.EUCompliance {
		req.EUCompliance = true
	}

	requestURL := cl.buildSearchURL(req)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	cl.setSearchHeaders(httpReq, req)

	client := &http.Client{}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return cl.parseSearchResponse(body, req.JSONResponse)
}

func (cl *Client) buildSearchURL(args SearchRequest) string {
	baseURL := "https://s.jina.ai/"
	if args.EUCompliance {
		baseURL = "https://eu.s.jina.ai/"
	}
	return baseURL
}

func (cl *Client) setSearchHeaders(req *http.Request, args SearchRequest) {
	req.Header.Add("Authorization", "Bearer "+cl.cfg.APIKey)

	if args.JSONResponse {
		req.Header.Add("Accept", "application/json")
	}

	if args.Site != "" {
		req.Header.Add("X-Site", args.Site)
	}
	if args.WithLinksSummary {
		req.Header.Add("X-With-Links-Summary", "true")
	}
	if args.WithImagesSummary {
		req.Header.Add("X-With-Images-Summary", "true")
	}
	if args.RetainImages != "" {
		req.Header.Add("X-Retain-Images", args.RetainImages)
	}
	if args.NoCache {
		req.Header.Add("X-No-Cache", "true")
	}
	if args.WithGeneratedAlt {
		req.Header.Add("X-With-Generated-Alt", "true")
	}
	if args.RespondWith != "" {
		req.Header.Add("X-Respond-With", args.RespondWith)
	}
	if args.WithFavicon {
		req.Header.Add("X-With-Favicon", "true")
	}
	if args.ReturnFormat != "" {
		req.Header.Add("X-Return-Format", args.ReturnFormat)
	}
	if args.Engine != "" {
		req.Header.Add("X-Engine", args.Engine)
	}
	if args.WithFavicons {
		req.Header.Add("X-With-Favicons", "true")
	}
	if args.Timeout > 0 {
		req.Header.Add("X-Timeout", strconv.Itoa(args.Timeout))
	}
	if args.SetCookie != "" {
		req.Header.Add("X-Set-Cookie", args.SetCookie)
	}
	if args.ProxyURL != "" {
		req.Header.Add("X-Proxy-Url", args.ProxyURL)
	}
	if args.Locale != "" {
		req.Header.Add("X-Locale", args.Locale)
	}
}

func (cl *Client) parseSearchResponse(body []byte, jsonResponse bool) (*SearchResponse, error) {
	if jsonResponse {
		var structured StructuredSearchResponse
		err := json.Unmarshal(body, &structured)
		if err != nil {
			return nil, fmt.Errorf("unmarshal response body: %w", err)
		}
		return &SearchResponse{Structured: &structured}, nil
	}
	return &SearchResponse{Text: string(body)}, nil
}
