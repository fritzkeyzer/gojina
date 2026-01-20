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

type BrowserEngine string

const (
	// BrowserEngineDefault Most compatible engine offering good balance between quality and speed.
	BrowserEngineDefault BrowserEngine = ""

	// BrowserEngineExperimental Fast engine capable of handling JavaScript-heavy websites we are
	// experimenting with. Not recommended for production use. We may change its behaviour.
	BrowserEngineExperimental BrowserEngine = "cf-browser-rendering"

	// BrowserEngineSpeed Fastest engine, optimized for speed but unable to handle JavaScript generated
	// dynamic content.
	BrowserEngineSpeed BrowserEngine = "direct"

	// BrowserEngineQuality High-quality engine designed to resolve rendering issues and deliver the best
	// content output.
	BrowserEngineQuality BrowserEngine = "browser"
)

type ContentFormat string

const (
	ContentFormatDefault    ContentFormat = ""           // The default pipeline optimized for most websites and LLM input
	ContentFormatMarkdown   ContentFormat = "markdown"   // Returns markdown directly from the HTML, bypassing the readability filtering
	ContentFormatHTML       ContentFormat = "html"       // Returns documentElement.outerHTML
	ContentFormatText       ContentFormat = "text"       // Returns document.body.innerText
	ContentFormatScreenshot ContentFormat = "screenshot" // Returns the image URL of the first screen
	ContentFormatPageshot   ContentFormat = "pageshot"   // Returns the image URL of the full page screenshot
)

type ReaderRequest struct {
	// URL is the URL to read and extract content from.
	URL string `json:"url"`

	// Viewport sets browser viewport dimensions for responsive rendering.
	Viewport *Viewport `json:"viewport,omitempty"`

	// InjectPageScript executes preprocessing JS code (inline string or remote URL) to manipulate DOMs.
	InjectPageScript string `json:"injectPageScript,omitempty"`

	// Header Options (mapped to HTTP headers)

	// JSONResponse if true, use application/json to get JSON response, text/event-stream to enable stream mode.
	JSONResponse bool `json:"-"`

	// BrowserEngine specifies the engine to retrieve/parse content. Use browser for fetching best quality content, direct for speed, cf-browser-rendering for experimental engine aimed at JS-heavy websites.
	BrowserEngine BrowserEngine `json:"-"`

	// ContentFormat specifies the return format: markdown, html, text, screenshot, or pageshot (for URL of full-page screenshot).
	ContentFormat ContentFormat `json:"-"`

	// Timeout specifies the maximum time (in seconds) to wait for the webpage to load.
	Timeout int `json:"-"`

	// TargetSelector CSS selectors to focus on specific elements within the page.
	TargetSelector string `json:"-"`

	// WaitForSelector CSS selectors to wait for specific elements before returning.
	WaitForSelector string `json:"-"`

	// RemoveSelector CSS selectors to exclude certain parts of the page (e.g., headers, footers).
	RemoveSelector string `json:"-"`

	// GatherLinks all to gather all links or true to gather unique links at the end of the response.
	GatherLinks string `json:"-"`

	// GatherImages all to gather all images or true to gather unique images at the end of the response.
	GatherImages string `json:"-"`

	// ImageCaption true to add alt text to images lacking captions.
	ImageCaption bool `json:"-"`

	// BypassCachedContent true to bypass cache for fresh retrieval.
	BypassCachedContent bool `json:"-"`

	// WithIframe true to include iframe content in the response.
	WithIframe bool `json:"-"`

	// TokenBudget specifies maximum number of tokens to use for the request.
	TokenBudget int `json:"-"`

	// RemoveAllImages use none to remove all images from the response.
	RemoveAllImages bool `json:"-"`

	// RespondWith use readerlm-v2, the language model specialized in HTML-to-Markdown, to deliver high-quality results for websites with complex structures and contents.
	RespondWith string `json:"-"`

	// SetCookie forwards your custom cookie settings when accessing the URL, which is useful for pages requiring extra authentication. Note that requests with cookies will not be cached.
	SetCookie string `json:"-"`

	// ProxyURL utilizes your proxy to access URLs, which is helpful for pages accessible only through specific proxies.
	ProxyURL string `json:"-"`

	// ProxyCountry sets country code for location-based proxy server. Use 'auto' for optimal selection or 'none' to disable.
	ProxyCountry string `json:"-"`

	// DNT use 1 to not cache and track the requested URL on our server.
	DNT int `json:"-"`

	// NoGfm opt in/out features from GFM (Github Flavored Markdown). By default, GFM (Github Flavored Markdown) features are enabled. Use true to disable GFM (Github Flavored Markdown) features. Use table to Opt out GFM Table but keep the table HTML elements in response.
	NoGfm string `json:"-"`

	// BrowserLocale controls the browser locale to render the page. Lots of websites serve different content based on the locale.
	BrowserLocale string `json:"-"`

	// RobotsTxt defines bot User-Agent to check against robots.txt before fetching content. Websites may allow different behaviors based on the User-Agent.
	RobotsTxt string `json:"-"`

	// WithShadowDom use true to extract content from all Shadow DOM roots in the document.
	WithShadowDom bool `json:"-"`

	// Base use final to follow the full redirect chain.
	Base string `json:"-"`

	// MdHeadingStyle when to use '#' or '===' to create Markdown headings. Set atx to use any number of \"==\" or \"--\" characters on the line below the text to create headings.
	MdHeadingStyle string `json:"-"`

	// MdHr defines Markdown horizontal rule format (passed to Turndown). Default is \"***\".
	MdHr string `json:"-"`

	// MdBulletListMarker sets Markdown bullet list marker character (passed to Turndown). Options: *, -, +
	MdBulletListMarker string `json:"-"`

	// MdEmDelimiter defines Markdown emphasis delimiter (passed to Turndown). Options: -, *
	MdEmDelimiter string `json:"-"`

	// MdStrongDelimiter sets Markdown strong emphasis delimiter (passed to Turndown). Options: **, __
	MdStrongDelimiter string `json:"-"`

	// MdLinkStyle when not set, links are embedded directly within the text. Sets referenced to list links at the end, referenced by numbers in the text. Sets discarded to replace links with their anchor text.
	MdLinkStyle string `json:"-"`

	// MdLinkReferenceStyle sets Markdown reference link format (passed to Turndown). Set to collapse, shortcut or do not set this header.
	MdLinkReferenceStyle string `json:"-"`

	// EUCompliance use EU infrastructure (eu.r.jina.ai) to reside all infrastructure and data processing operations entirely within EU jurisdiction.
	EUCompliance bool `json:"-"`
}

type Viewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type ReaderResponse struct {
	Text       string                    // Raw text response (when JSON is not requested)
	Structured *StructuredReaderResponse // Structured JSON response
}

type StructuredReaderResponse struct {
	Code   int `json:"code"`
	Status int `json:"status"`
	Data   struct {
		Title       string            `json:"title"`
		Description string            `json:"description"`
		URL         string            `json:"url"`
		Content     string            `json:"content"`
		Links       map[string]string `json:"links,omitempty"`
		Images      map[string]string `json:"images,omitempty"`
		Usage       struct {
			Tokens int `json:"tokens"`
		} `json:"usage"`
	} `json:"data"`
}

// Reader calls the Jina Reader API to retrieve and parse content from a URL.
func (cl *Client) Reader(ctx context.Context, req ReaderRequest) (*ReaderResponse, error) {
	if req.URL == "" {
		return nil, fmt.Errorf("URL is required")
	}
	if cl.cfg.EUCompliance {
		req.EUCompliance = true
	}

	requestURL := cl.buildReaderURL(req)

	// Marshal only the body parameters
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	cl.setReaderHeaders(httpReq, req)

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

	return cl.parseReaderResponse(body, req.JSONResponse)
}

func (cl *Client) buildReaderURL(args ReaderRequest) string {
	baseURL := "https://r.jina.ai/"
	if args.EUCompliance {
		baseURL = "https://eu.r.jina.ai/"
	}

	return baseURL
}

func (cl *Client) setReaderHeaders(httpReq *http.Request, req ReaderRequest) {
	httpReq.Header.Add("Authorization", "Bearer "+cl.cfg.APIKey)
	if req.TokenBudget > 0 {
		httpReq.Header.Add("X-Token-Budget", fmt.Sprintf("%d", req.TokenBudget))
	}

	if req.ContentFormat != ContentFormatDefault {
		httpReq.Header.Add("X-Return-Format", string(req.ContentFormat))
	}

	if req.BrowserEngine != BrowserEngineDefault {
		httpReq.Header.Add("X-Engine", string(req.BrowserEngine))
	}

	if req.Timeout > 0 {
		httpReq.Header.Add("X-Timeout", strconv.Itoa(req.Timeout))
	}

	if req.GatherLinks != "" {
		httpReq.Header.Add("X-With-Links-Summary", req.GatherLinks)
	}

	if req.RemoveAllImages {
		httpReq.Header.Add("X-Retain-Images", "none")
	}

	if req.GatherImages != "" {
		httpReq.Header.Add("X-With-Images-Summary", req.GatherImages)
	}

	if req.ImageCaption {
		httpReq.Header.Add("X-With-Generated-Alt", "true")
	}

	if req.ProxyCountry != "" {
		httpReq.Header.Add("X-Proxy", req.ProxyCountry)
	}

	if req.ProxyURL != "" {
		httpReq.Header.Add("X-Proxy-Url", req.ProxyURL)
	}

	if req.BrowserLocale != "" {
		httpReq.Header.Add("X-Locale", req.BrowserLocale)
	}

	if req.BypassCachedContent {
		httpReq.Header.Add("X-No-Cache", "true")
	}

	if req.TargetSelector != "" {
		httpReq.Header.Add("X-Target-Selector", req.TargetSelector)
	}

	if req.WaitForSelector != "" {
		httpReq.Header.Add("X-Wait-For-Selector", req.WaitForSelector)
	}

	if req.RemoveSelector != "" {
		httpReq.Header.Add("X-Remove-Selector", req.RemoveSelector)
	}

	if req.WithIframe {
		httpReq.Header.Add("X-With-Iframe", "true")
	}

	if req.WithShadowDom {
		httpReq.Header.Add("X-With-Shadow-Dom", "true")
	}

	if req.RespondWith != "" {
		httpReq.Header.Add("X-Respond-With", req.RespondWith)
	}

	if req.SetCookie != "" {
		httpReq.Header.Add("X-Set-Cookie", req.SetCookie)
	}

	if req.DNT > 0 {
		httpReq.Header.Add("DNT", strconv.Itoa(req.DNT))
	}

	if req.NoGfm != "" {
		httpReq.Header.Add("X-No-Gfm", req.NoGfm)
	}

	if req.RobotsTxt != "" {
		httpReq.Header.Add("X-Robots-Txt", req.RobotsTxt)
	}

	if req.Base != "" {
		httpReq.Header.Add("X-Base", req.Base)
	}

	if req.MdHeadingStyle != "" {
		httpReq.Header.Add("X-Md-Heading-Style", req.MdHeadingStyle)
	}

	if req.MdHr != "" {
		httpReq.Header.Add("X-Md-Hr", req.MdHr)
	}

	if req.MdBulletListMarker != "" {
		httpReq.Header.Add("X-Md-Bullet-List-Marker", req.MdBulletListMarker)
	}

	if req.MdEmDelimiter != "" {
		httpReq.Header.Add("X-Md-Em-Delimiter", req.MdEmDelimiter)
	}

	if req.MdStrongDelimiter != "" {
		httpReq.Header.Add("X-Md-Strong-Delimiter", req.MdStrongDelimiter)
	}

	if req.MdLinkStyle != "" {
		httpReq.Header.Add("X-Md-Link-Style", req.MdLinkStyle)
	}

	if req.MdLinkReferenceStyle != "" {
		httpReq.Header.Add("X-Md-Link-Reference-Style", req.MdLinkReferenceStyle)
	}

	if req.JSONResponse {
		httpReq.Header.Add("Accept", "application/json")
	}
}

func (cl *Client) parseReaderResponse(body []byte, jsonResponse bool) (*ReaderResponse, error) {
	if jsonResponse {
		var structured StructuredReaderResponse

		err := json.Unmarshal(body, &structured)
		if err != nil {
			return nil, fmt.Errorf("unmarshal response body: %w", err)
		}

		return &ReaderResponse{
			Structured: &structured,
		}, nil
	}

	return &ReaderResponse{
		Text:       string(body),
		Structured: nil,
	}, nil
}
