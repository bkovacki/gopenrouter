package gopenrouter

import (
	"context"
	"net/http"
)

// modelsResponse represents the internal API response structure when listing models.
// It wraps the actual model data in a 'data' field.
type modelsResponse struct {
	Data []ModelData `json:"data"`
}

// ModelData represents information about an AI model available through OpenRouter.
// It contains details about the model's capabilities, pricing, and technical specifications.
type ModelData struct {
	// ID is the unique identifier for the model
	ID string `json:"id"`
	// Name is the human-readable name of the model
	Name string `json:"name"`
	// Created is the Unix timestamp when the model was added to OpenRouter
	Created float32 `json:"created"`
	// Description provides details about the model's capabilities
	Description string `json:"description"`
	// Architecture contains information about the model's input/output capabilities
	Architecture ModelArchitecture `json:"architecture"`
	// TopProvider contains information about the primary provider for this model
	TopProvider ModelTopProvider `json:"top_provider"`
	// Pricing contains the cost information for using this model
	Pricing ModelPricing `json:"pricing"`
	// ContextLength is the maximum number of tokens the model can process
	ContextLength float32 `json:"context_length,omitempty"`
	// PerRequestLimits contains any limitations on requests to this model
	PerRequestLimits map[string]any `json:"per_request_limits,omitempty"`
	// SupportedParameters lists all parameters that can be used with this model
	// Note: This is a union of parameters from all providers; no single provider may support all parameters
	SupportedParameters []string `json:"supported_parameters,omitempty"`
}

// ModelArchitecture contains information about the model's input and output capabilities.
type ModelArchitecture struct {
	// InputModalities describes the types of input the model can accept (e.g., "text", "image")
	InputModalities []string `json:"input_modalities"`
	// OutputModalities describes the types of output the model can produce (e.g., "text")
	OutputModalities []string `json:"output_modalities"`
	// Tokenizer indicates the tokenization method used by the model (e.g., "GPT")
	Tokenizer string `json:"tokenizer"`
	// InstructType specifies the instruction format the model uses (if applicable)
	InstructType string `json:"instruct_type,omitempty"`
}

// ModelTopProvider contains information about the primary provider for a model.
type ModelTopProvider struct {
	// IsModerated indicates if the provider applies content moderation
	IsModerated bool `json:"is_moderated"`
	// ContextLength is the maximum context length supported by this specific provider
	ContextLength float32 `json:"context_length,omitempty"`
	// MaxCompletionTokent is the maximum number of tokens the provider allows in completions
	// Note: Field name has a typo but matches the API response
	MaxCompletionTokent float32 `json:"max_completion_tokens,omitempty"`
}

// ModelPricing contains the cost information for using a model.
// All prices are expressed as strings representing cost per token (or per operation).
type ModelPricing struct {
	// Prompt is the cost per token for the input/prompt
	Prompt string `json:"prompt"`
	// Completion is the cost per token for the output/completion
	Completion string `json:"completion"`
	// Image is the cost per image in the input
	Image string `json:"image"`
	// Request is the fixed cost per request
	Request string `json:"request"`
	// InputCacheRead is the cost for reading from the prompt cache
	InputCacheRead string `json:"input_cache_read"`
	// InputCacheWrite is the cost for writing to the prompt cache
	InputCacheWrite string `json:"input_cache_write"`
	// WebSearch is the cost for web search operations
	WebSearch string `json:"web_search"`
	// InternalReasoning is the cost for internal reasoning tokens
	InternalReasoning string `json:"internal_reasoning"`
}

// ListModels retrieves information about all models available through the OpenRouter API.
//
// The returned list includes details about each model's capabilities, pricing,
// and technical specifications. This information can be used to select an appropriate
// model for different use cases or to compare models.
//
// Parameters:
//   - ctx: The context for the request, which can be used for cancellation and timeouts
//
// Returns:
//   - []ModelData: A list of available models with their details
//   - error: Any error that occurred during the request
func (c *Client) ListModels(ctx context.Context) (models []ModelData, err error) {
	var response modelsResponse
	urlSuffix := "/models"

	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	if err != nil {
		return
	}

	models = response.Data
	return
}
