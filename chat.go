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

// ChatCompletionRequest represents a request for chat completion to the OpenRouter API.
// It contains all the parameters needed to generate chat responses from AI models.
type ChatCompletionRequest struct {
	// Required fields
	// Model is the identifier of the AI model to use
	Model string `json:"model"`
	// Messages is the conversation history as a list of messages
	Messages []ChatMessage `json:"messages"`

	// Optional fields
	// Models provides an alternate list of models for routing overrides
	Models []string `json:"models,omitempty"`
	// Provider contains preferences for provider routing
	Provider *ProviderOptions `json:"provider,omitempty"`
	// Reasoning configures model reasoning/thinking tokens
	Reasoning *ReasoningOptions `json:"reasoning,omitempty"`
	// Usage specifies whether to include usage information in the response
	Usage *UsageOptions `json:"usage,omitempty"`
	// Transforms lists prompt transformations (OpenRouter-only feature)
	Transforms []string `json:"transforms,omitempty"`
	// Stream enables streaming of results as they are generated
	Stream *bool `json:"stream,omitempty"`
	// MaxTokens limits the maximum number of tokens in the response
	MaxTokens *int `json:"max_tokens,omitempty"`
	// Temperature controls randomness in generation (range: [0, 2])
	Temperature *float64 `json:"temperature,omitempty"`
	// Seed enables deterministic outputs with the same inputs
	Seed *int `json:"seed,omitempty"`
	// TopP controls nucleus sampling (range: (0, 1])
	TopP *float64 `json:"top_p,omitempty"`
	// TopK limits sampling to top K most likely tokens (range: [1, Infinity))
	TopK *int `json:"top_k,omitempty"`
	// FrequencyPenalty reduces repetition of token sequences (range: [-2, 2])
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	// PresencePenalty reduces repetition of topics (range: [-2, 2])
	PresencePenalty *float64 `json:"presence_penalty,omitempty"`
	// RepetitionPenalty penalizes repeated tokens (range: (0, 2])
	RepetitionPenalty *float64 `json:"repetition_penalty,omitempty"`
	// LogitBias maps token IDs to bias values for controlling token probability
	LogitBias map[string]float64 `json:"logit_bias,omitempty"`
	// TopLogProbs specifies the number of top log probabilities to return
	TopLogProbs *int `json:"top_logprobs,omitempty"`
	// MinP sets the minimum probability threshold for tokens (range: [0, 1])
	MinP *float64 `json:"min_p,omitempty"`
	// TopA is an alternate top sampling parameter (range: [0, 1])
	TopA *float64 `json:"top_a,omitempty"`
	// Logprobs enables returning log probabilities of output tokens
	Logprobs *bool `json:"logprobs,omitempty"`
	// Stop specifies sequences where the model will stop generating tokens
	Stop []string `json:"stop,omitempty"`
	// User is a stable identifier for end-users, used to help detect and prevent abuse
	User *string `json:"user,omitempty"`
}

// ChatMessage represents a single message in a conversation.
// Each message has a role (system, user, assistant) and content.
type ChatMessage struct {
	// Role defines who sent the message (system, user, or assistant)
	Role string `json:"role"`
	// Content is the text content of the message
	Content string `json:"content"`
}

// ChatCompletionResponse represents the response from a chat completion request.
// It contains the generated messages and metadata about the request.
type ChatCompletionResponse struct {
	// ID is the unique identifier for this chat completion request
	ID string `json:"id"`
	// Choices contains the generated chat message responses
	Choices []ChatChoice `json:"choices"`
	// Usage provides token usage statistics for the request
	Usage Usage `json:"usage,omitzero"`
}

// ChatChoice represents a single chat completion choice from the API.
// The API may return multiple choices depending on the request parameters.
type ChatChoice struct {
	// Message is the generated chat message response
	Message ChatMessage `json:"message"`
	// Index is the position of this choice in the array of choices
	Index int `json:"index,omitempty"`
	// FinishReason explains why the generation stopped (e.g., "stop", "length")
	FinishReason string `json:"finish_reason,omitempty"`
}

// ChatCompletionRequestBuilder implements a builder pattern for constructing ChatCompletionRequest objects.
// It provides a fluent interface for setting request parameters with method chaining.
type ChatCompletionRequestBuilder struct {
	request *ChatCompletionRequest
}

// NewChatCompletionRequestBuilder creates a new builder for ChatCompletionRequest with required fields.
// The model and messages parameters are required for all chat completion requests.
func NewChatCompletionRequestBuilder(model string, messages []ChatMessage) *ChatCompletionRequestBuilder {
	return &ChatCompletionRequestBuilder{
		request: &ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	}
}

// WithModels sets alternate models for routing overrides.
func (b *ChatCompletionRequestBuilder) WithModels(models []string) *ChatCompletionRequestBuilder {
	b.request.Models = models
	return b
}

// WithProvider sets provider preferences for routing.
func (b *ChatCompletionRequestBuilder) WithProvider(provider *ProviderOptions) *ChatCompletionRequestBuilder {
	b.request.Provider = provider
	return b
}

// WithReasoning sets reasoning configuration for the request.
func (b *ChatCompletionRequestBuilder) WithReasoning(reasoning *ReasoningOptions) *ChatCompletionRequestBuilder {
	b.request.Reasoning = reasoning
	return b
}

// WithUsage sets whether to include usage information in the response.
func (b *ChatCompletionRequestBuilder) WithUsage(include bool) *ChatCompletionRequestBuilder {
	b.request.Usage = &UsageOptions{
		Include: &include,
	}
	return b
}

// WithTransforms sets prompt transformations for the request.
func (b *ChatCompletionRequestBuilder) WithTransforms(transforms []string) *ChatCompletionRequestBuilder {
	b.request.Transforms = transforms
	return b
}

// WithStream enables or disables streaming for the request.
func (b *ChatCompletionRequestBuilder) WithStream(stream bool) *ChatCompletionRequestBuilder {
	b.request.Stream = &stream
	return b
}

// WithMaxTokens sets the maximum number of tokens for the response.
func (b *ChatCompletionRequestBuilder) WithMaxTokens(maxTokens int) *ChatCompletionRequestBuilder {
	b.request.MaxTokens = &maxTokens
	return b
}

// WithTemperature sets the sampling temperature for the request.
func (b *ChatCompletionRequestBuilder) WithTemperature(temperature float64) *ChatCompletionRequestBuilder {
	b.request.Temperature = &temperature
	return b
}

// WithSeed sets the seed for deterministic outputs.
func (b *ChatCompletionRequestBuilder) WithSeed(seed int) *ChatCompletionRequestBuilder {
	b.request.Seed = &seed
	return b
}

// WithTopP sets the nucleus sampling parameter.
func (b *ChatCompletionRequestBuilder) WithTopP(topP float64) *ChatCompletionRequestBuilder {
	b.request.TopP = &topP
	return b
}

// WithTopK sets the top-k sampling parameter.
func (b *ChatCompletionRequestBuilder) WithTopK(topK int) *ChatCompletionRequestBuilder {
	b.request.TopK = &topK
	return b
}

// WithFrequencyPenalty sets the frequency penalty parameter.
func (b *ChatCompletionRequestBuilder) WithFrequencyPenalty(penalty float64) *ChatCompletionRequestBuilder {
	b.request.FrequencyPenalty = &penalty
	return b
}

// WithPresencePenalty sets the presence penalty parameter.
func (b *ChatCompletionRequestBuilder) WithPresencePenalty(penalty float64) *ChatCompletionRequestBuilder {
	b.request.PresencePenalty = &penalty
	return b
}

// WithRepetitionPenalty sets the repetition penalty parameter.
func (b *ChatCompletionRequestBuilder) WithRepetitionPenalty(penalty float64) *ChatCompletionRequestBuilder {
	b.request.RepetitionPenalty = &penalty
	return b
}

// WithLogitBias sets the logit bias for specific tokens.
func (b *ChatCompletionRequestBuilder) WithLogitBias(logitBias map[string]float64) *ChatCompletionRequestBuilder {
	b.request.LogitBias = logitBias
	return b
}

// WithTopLogprobs sets the number of top log probabilities to return.
func (b *ChatCompletionRequestBuilder) WithTopLogprobs(topLogProbs int) *ChatCompletionRequestBuilder {
	b.request.TopLogProbs = &topLogProbs
	return b
}

// WithMinP sets the minimum probability threshold.
func (b *ChatCompletionRequestBuilder) WithMinP(minP float64) *ChatCompletionRequestBuilder {
	b.request.MinP = &minP
	return b
}

// WithTopA sets the top-a sampling parameter.
func (b *ChatCompletionRequestBuilder) WithTopA(topA float64) *ChatCompletionRequestBuilder {
	b.request.TopA = &topA
	return b
}

// WithLogprobs enables or disables returning log probabilities of output tokens.
func (b *ChatCompletionRequestBuilder) WithLogprobs(logprobs bool) *ChatCompletionRequestBuilder {
	b.request.Logprobs = &logprobs
	return b
}

// WithStop sets the stop sequences for token generation.
func (b *ChatCompletionRequestBuilder) WithStop(stop []string) *ChatCompletionRequestBuilder {
	b.request.Stop = stop
	return b
}

// WithUser sets the user identifier for the request.
func (b *ChatCompletionRequestBuilder) WithUser(user string) *ChatCompletionRequestBuilder {
	b.request.User = &user
	return b
}

// Build returns the constructed ChatCompletionRequest.
func (b *ChatCompletionRequestBuilder) Build() *ChatCompletionRequest {
	return b.request
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

// ChatCompletionStreamResponse represents a single chunk in a streaming chat completion response
type ChatCompletionStreamResponse struct {
	ID      string                `json:"id"`
	Object  string                `json:"object"`
	Created int64                 `json:"created"`
	Model   string                `json:"model"`
	Choices []ChatStreamingChoice `json:"choices"`
	Usage   *Usage                `json:"usage,omitempty"`
}

// ChatCompletionStreamReader implements StreamReader for chat completion responses
type ChatCompletionStreamReader struct {
	reader   *bufio.Scanner
	response *http.Response
	buffer   string
}

// NewChatCompletionStreamReader creates a new stream reader for chat completion responses
func NewChatCompletionStreamReader(response *http.Response) *ChatCompletionStreamReader {
	return &ChatCompletionStreamReader{
		reader:   bufio.NewScanner(response.Body),
		response: response,
	}
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

// ChatCompletion sends a chat completion request to the OpenRouter API.
//
// This method allows users to generate chat responses from AI models through the
// OpenRouter API. The request can be customized with various parameters to control
// the generation process, provider selection, and output format.
//
// The method takes a context for cancellation and timeout control, and a ChatCompletionRequest
// containing the conversation messages and generation parameters.
//
// Returns a ChatCompletionResponse containing the generated messages and usage statistics,
// or an error if the request fails.
func (c *Client) ChatCompletion(
	ctx context.Context,
	request ChatCompletionRequest,
) (response ChatCompletionResponse, err error) {
	if request.Stream != nil && *request.Stream {
		err = ErrCompletionStreamNotSupported
		return
	}

	urlSuffix := "/chat/completions"

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix),
		withBody(request),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
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
