package gopenrouter

import (
	"context"
	"fmt"
	"net/http"
)

// endpointsResponse represents the internal API response when retrieving endpoints for a model.
// It wraps the endpoint data in a standard response structure.
type endpointsResponse struct {
	Data EndpointData `json:"data"`
}

// EndpointData contains information about a model and its available endpoints.
// This includes both model metadata and a list of provider-specific endpoints.
type EndpointData struct {
	// ID is the unique identifier for the model
	ID string `json:"id"`
	// Name is the human-readable name of the model
	Name string `json:"name"`
	// Created is the Unix timestamp when the model was added
	Created float64 `json:"created"`
	// Description provides details about the model's capabilities
	Description string `json:"description"`
	// Architecture contains information about the model's input/output capabilities
	Architecture Architecture `json:"architecture"`
	// Endpoints is a list of provider-specific implementations of this model
	Endpoints []EndpointDetail `json:"endpoints"`
}

// Architecture describes a model's input and output capabilities.
type Architecture struct {
	// InputModalities lists the types of input the model accepts (e.g., "text", "image")
	InputModalities []string `json:"input_modalities"`
	// OutputModalities lists the types of output the model produces (e.g., "text")
	OutputModalities []string `json:"output_modalities"`
	// Tokenizer indicates the tokenization method used by this model
	Tokenizer string `json:"tokenizer"`
	// InstructType specifies the instruction format for this model
	InstructType string `json:"instruct_type"`
}

// EndpointDetail represents a specific provider endpoint for a model.
// Each endpoint is a provider-specific implementation of the same model.
type EndpointDetail struct {
	// Name is the identifier for this specific endpoint
	Name string `json:"name"`
	// ContextLength is the maximum context length supported by this endpoint
	ContextLength float64 `json:"context_length"`
	// Pricing contains the cost information for using this endpoint
	Pricing EndpointPricing `json:"pricing"`
	// ProviderName identifies which AI provider offers this endpoint
	ProviderName string `json:"provider_name"`
	// SupportedParameters lists the API parameters this endpoint accepts
	SupportedParameters []string `json:"supported_parameters"`
}

// EndpointPricing contains pricing information for using a specific endpoint.
// All prices are expressed as strings representing cost per token (or per operation).
type EndpointPricing struct {
	// Request is the fixed cost per API request
	Request string `json:"request"`
	// Image is the cost per image in the input
	Image string `json:"image"`
	// Prompt is the cost per token for the input/prompt
	Prompt string `json:"prompt"`
	// Completion is the cost per token for the output/completion
	Completion string `json:"completion"`
}

// ListEndpoints retrieves information about all available endpoints for a specific model.
//
// Each model on OpenRouter may be available through multiple providers, with each provider
// offering different capabilities, context lengths, and pricing. This method allows you
// to see all provider-specific implementations of a given model.
//
// Parameters:
//   - ctx: The context for the request, which can be used for cancellation and timeouts
//   - author: The author/owner of the model
//   - slug: The model identifier/slug
//
// Returns:
//   - EndpointData: Contains model information and a list of available endpoints
//   - error: Any error that occurred during the request
func (c *Client) ListEndpoints(ctx context.Context, author string, slug string) (data EndpointData, err error) {
	urlSuffix := fmt.Sprintf("/models/%s/%s/endpoints", author, slug)
	var response endpointsResponse

	req, err := c.newRequest(ctx, http.MethodGet, c.fullURL(urlSuffix))
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
