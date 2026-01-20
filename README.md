# gojina

gojina is a Go client library for interacting with [Jina AI's Search Foundation APIs](https://jina.ai/).

It provides a simple and strongly-typed interface for:
- **Reader API**: Extract content from URLs.
- **Embeddings API**: Generate text and multimodal embeddings.
- **Reranker API**: Re-rank search results for better relevance.
- **Search API**: Search the web with LLM-friendly output.
- **DeepSearch API**: Complex reasoning and web investigation.
- **VLM API**: Vision Language Models for image understanding.
- **Classification API**: Classify text and images.
- **Segmenter API**: Tokenize and chunk text.

## Installation

```bash
go get github.com/fritzkeyzer/gojina
```

## Usage

You need a Jina AI API key to use most of these services. Get one for free at [jina.ai](https://jina.ai/?sui=apikey).

For comprehensive examples of how to use each API, please refer to the [examples](./examples) directory.
Eg: [search](./examples/search/main.go) or [reader](./examples/reader/main.go)

Each API has its own example showing how to construct requests and handle responses.
