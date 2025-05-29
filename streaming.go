package gopenrouter

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// StreamingChoice represents a streaming completion choice with text content
type StreamingChoice struct {
	Index              int     `json:"index"`
	Text               string  `json:"text"`
	FinishReason       *string `json:"finish_reason"`
	NativeFinishReason *string `json:"native_finish_reason"`
	Logprobs           *string `json:"logprobs"`
}

// ChatStreamingChoice represents a streaming chat completion choice with delta content
type ChatStreamingChoice struct {
	Index        int       `json:"index"`
	Delta        ChatDelta `json:"delta"`
	FinishReason *string   `json:"finish_reason"`
}

// ChatDelta represents the incremental content in a streaming chat response
type ChatDelta struct {
	Role    *string `json:"role,omitempty"`
	Content *string `json:"content,omitempty"`
}

// CompletionStreamResponse represents a single chunk in a streaming completion response
type CompletionStreamResponse struct {
	ID               string            `json:"id"`
	Provider         string            `json:"provider"`
	Model            string            `json:"model"`
	Object           string            `json:"object"`
	Created          int64             `json:"created"`
	Choices          []StreamingChoice `json:"choices"`
	SystemFingerprint *string          `json:"system_fingerprint,omitempty"`
	Usage            *Usage            `json:"usage,omitempty"`
}

// ChatCompletionStreamResponse represents a single chunk in a streaming chat completion response
type ChatCompletionStreamResponse struct {
	ID      string                `json:"id"`
	Object  string                `json:"object"`
	Created int64                 `json:"created"`
	Model   string                `json:"model"`
	Choices []ChatStreamingChoice `json:"choices"`
	Usage   *Usage                `json:"usage,omitempty"`
}

// StreamReader represents a generic interface for reading streaming responses
type StreamReader[T any] interface {
	// Recv reads the next chunk from the stream
	Recv() (T, error)
	// Close closes the stream and cleans up resources
	Close() error
}

// CompletionStreamReader implements StreamReader for completion responses
type CompletionStreamReader struct {
	reader   *bufio.Scanner
	response *http.Response
	buffer   string
}

// ChatCompletionStreamReader implements StreamReader for chat completion responses
type ChatCompletionStreamReader struct {
	reader   *bufio.Scanner
	response *http.Response
	buffer   string
}

// NewCompletionStreamReader creates a new stream reader for completion responses
func NewCompletionStreamReader(response *http.Response) *CompletionStreamReader {
	return &CompletionStreamReader{
		reader:   bufio.NewScanner(response.Body),
		response: response,
	}
}

// NewChatCompletionStreamReader creates a new stream reader for chat completion responses
func NewChatCompletionStreamReader(response *http.Response) *ChatCompletionStreamReader {
	return &ChatCompletionStreamReader{
		reader:   bufio.NewScanner(response.Body),
		response: response,
	}
}

// Recv reads the next completion chunk from the stream
func (r *CompletionStreamReader) Recv() (CompletionStreamResponse, error) {
	var response CompletionStreamResponse

	for {
		if !r.reader.Scan() {
			if err := r.reader.Err(); err != nil {
				return response, fmt.Errorf("error reading stream: %w", err)
			}
			return response, io.EOF
		}

		line := strings.TrimSpace(r.reader.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse SSE data
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// Check for stream end
			if data == "[DONE]" {
				return response, io.EOF
			}

			// Parse JSON chunk
			if err := json.Unmarshal([]byte(data), &response); err != nil {
				// Skip malformed chunks
				continue
			}

			return response, nil
		}
	}
}

// Close closes the completion stream reader
func (r *CompletionStreamReader) Close() error {
	if r.response != nil && r.response.Body != nil {
		return r.response.Body.Close()
	}
	return nil
}

// Recv reads the next chat completion chunk from the stream
func (r *ChatCompletionStreamReader) Recv() (ChatCompletionStreamResponse, error) {
	var response ChatCompletionStreamResponse

	for {
		if !r.reader.Scan() {
			if err := r.reader.Err(); err != nil {
				return response, fmt.Errorf("error reading stream: %w", err)
			}
			return response, io.EOF
		}

		line := strings.TrimSpace(r.reader.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse SSE data
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// Check for stream end
			if data == "[DONE]" {
				return response, io.EOF
			}

			// Parse JSON chunk
			if err := json.Unmarshal([]byte(data), &response); err != nil {
				// Skip malformed chunks
				continue
			}

			return response, nil
		}
	}
}

// Close closes the chat completion stream reader
func (r *ChatCompletionStreamReader) Close() error {
	if r.response != nil && r.response.Body != nil {
		return r.response.Body.Close()
	}
	return nil
}

// CompletionStream sends a streaming completion request to the OpenRouter API.
//
// This method enables real-time streaming of completion responses, allowing applications
// to display partial results as they are generated by the AI model.
//
// The method automatically sets the stream parameter to true in the request and returns
// a CompletionStreamReader for reading the streaming chunks.
//
// Example usage:
//
//	request := gopenrouter.NewCompletionRequestBuilder("model-id", "prompt").Build()
//	stream, err := client.CompletionStream(ctx, *request)
//	if err != nil {
//	  // handle error
//	}
//	defer stream.Close()
//
//	for {
//	  chunk, err := stream.Recv()
//	  if err == io.EOF {
//	    break // Stream finished
//	  }
//	  if err != nil {
//	    // handle error
//	  }
//	  // Process chunk
//	}
func (c *Client) CompletionStream(
	ctx context.Context,
	request CompletionRequest,
) (*CompletionStreamReader, error) {
	// Ensure stream is enabled on a copy of the request
	streamEnabled := true
	request.Stream = &streamEnabled

	urlSuffix := "/completions"

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(request),
	)
	if err != nil {
		return nil, err
	}

	// Set accept header for streaming
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		defer resp.Body.Close()
		return nil, c.handleErrorResp(resp)
	}

	return NewCompletionStreamReader(resp), nil
}

// ChatCompletionStream sends a streaming chat completion request to the OpenRouter API.
//
// This method enables real-time streaming of chat completion responses, allowing applications
// to display partial results as they are generated by the AI model.
//
// The method automatically sets the stream parameter to true in the request and returns
// a ChatCompletionStreamReader for reading the streaming chunks.
//
// Example usage:
//
//	messages := []gopenrouter.ChatMessage{{Role: "user", Content: "Hello"}}
//	request := gopenrouter.NewChatCompletionRequestBuilder("model-id", messages).Build()
//	stream, err := client.ChatCompletionStream(ctx, *request)
//	if err != nil {
//	  // handle error
//	}
//	defer stream.Close()
//
//	for {
//	  chunk, err := stream.Recv()
//	  if err == io.EOF {
//	    break // Stream finished
//	  }
//	  if err != nil {
//	    // handle error
//	  }
//	  // Process chunk
//	}
func (c *Client) ChatCompletionStream(
	ctx context.Context,
	request ChatCompletionRequest,
) (*ChatCompletionStreamReader, error) {
	// Ensure stream is enabled
	streamEnabled := true
	request.Stream = &streamEnabled

	urlSuffix := "/chat/completions"

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(request),
	)
	if err != nil {
		return nil, err
	}

	// Set accept header for streaming
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		defer resp.Body.Close()
		return nil, c.handleErrorResp(resp)
	}

	return NewChatCompletionStreamReader(resp), nil
}
