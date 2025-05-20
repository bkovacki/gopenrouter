package gopenrouter

import (
	"context"
	"net/http"
)

// generationResponse represents the internal API response when retrieving a single generation's data.
// It wraps the generation data in a standard response structure.
type generationResponse struct {
	Data GenerationData `json:"data"`
}

// GenerationData contains detailed information about a specific generation request.
// This includes metadata about the request, the model used, performance metrics,
// token usage statistics, and other details about the generation process.
type GenerationData struct {
	// ID is the unique identifier for this generation
	ID string `json:"id"`
	// TotalCost represents the total cost of the generation in credits
	TotalCost float64 `json:"total_cost"`
	// CreatedAt is the timestamp when the generation was created
	CreatedAt string `json:"created_at"`
	// Model is the name of the AI model used for the generation
	Model string `json:"model"`
	// Origin indicates the source of the generation request
	Origin string `json:"origin"`
	// Usage represents the total credit usage for this generation
	Usage float64 `json:"usage"`
	// IsBYOK indicates if this was a "Bring Your Own Key" request
	IsBYOK bool `json:"is_byok"`
	// UpstreamID is the ID assigned by the upstream provider
	UpstreamID string `json:"upstream_id"`
	// CacheDiscount represents any discount applied due to prompt caching
	CacheDiscount float64 `json:"cache_discount"`
	// AppID is the identifier of the application that made the request
	AppID int `json:"app_id"`
	// Streamed indicates whether the generation was streamed
	Streamed bool `json:"streamed"`
	// Cancelled indicates whether the generation was cancelled before completion
	Cancelled bool `json:"cancelled"`
	// ProviderName is the name of the AI provider (e.g., "openai", "anthropic")
	ProviderName string `json:"provider_name"`
	// Latency is the total latency of the request in milliseconds
	Latency int `json:"latency"`
	// ModerationLatency is the time spent on content moderation in milliseconds
	ModerationLatency int `json:"moderation_latency"`
	// GenerationTime is the time spent generating the response in milliseconds
	GenerationTime int `json:"generation_time"`
	// FinishReason describes why the generation stopped
	FinishReason string `json:"finish_reason"`
	// NativeFinishReason is the raw finish reason from the provider
	NativeFinishReason string `json:"native_finish_reason"`
	// TokensPrompt is the number of tokens in the prompt
	TokensPrompt int `json:"tokens_prompt"`
	// TokensCompletion is the number of tokens in the completion
	TokensCompletion int `json:"tokens_completion"`
	// NativeTokensPrompt is the raw token count from the provider for the prompt
	NativeTokensPrompt int `json:"native_tokens_prompt"`
	// NativeTokensCompletion is the raw token count from the provider for the completion
	NativeTokensCompletion int `json:"native_tokens_completion"`
	// NativeTokensReasoning is the number of tokens used for internal reasoning
	NativeTokensReasoning int `json:"native_tokens_reasoning"`
	// NumMediaPrompt is the count of media items in the prompt
	NumMediaPrompt int `json:"num_media_prompt"`
	// NumMediaCompletion is the count of media items in the completion
	NumMediaCompletion int `json:"num_media_completion"`
	// NumSearchResults is the number of search results included
	NumSearchResults int `json:"num_search_results"`
}

// GetGeneration retrieves metadata about a specific generation request by its ID.
//
// The generation ID is provided when creating a completion or chat completion.
// This method allows you to retrieve detailed information about a previously
// made request, including its cost, token usage, and other metadata.
//
// Parameters:
//   - ctx: The context for the request, which can be used for cancellation and timeouts
//   - id: The unique identifier of the generation to retrieve
//
// Returns:
//   - GenerationData: Contains the detailed generation metadata
//   - error: Any error that occurred during the request
func (c *Client) GetGeneration(ctx context.Context, id string) (data GenerationData, err error) {
	urlSuffix := "/generation"
	var response generationResponse

	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(urlSuffix),
		withQueryParam("id", id),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	if err != nil {
		return
	}

	data = response.Data
	return
}
