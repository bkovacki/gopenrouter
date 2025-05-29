package gopenrouter

import (
	"context"
	"net/http"
)

// Effort represents the level of token allocation for reasoning in AI models.
// Different effort levels allocate different proportions of the maximum token limit.
type Effort string

const (
	// EffortHigh allocates a large portion of tokens for reasoning (approximately 80% of max_tokens)
	EffortHigh Effort = "high"

	// EffortMedium allocates a moderate portion of tokens for reasoning (approximately 50% of max_tokens)
	EffortMedium Effort = "medium"

	// EffortLow allocates a smaller portion of tokens for reasoning (approximately 20% of max_tokens)
	EffortLow Effort = "Low"
)

// Quantization represents the precision level used in model weights.
// Different quantization levels offer trade-offs between model size, inference speed,
// and prediction quality.
type Quantization string

const (
	// QuantizationInt4 represents Integer (4 bit) quantization
	QuantizationInt4 Quantization = "int4"

	// QuantizationInt8 represents Integer (8 bit) quantization
	QuantizationInt8 Quantization = "int8"

	// QuantizationFP4 represents Floating point (4 bit) quantization
	QuantizationFP4 Quantization = "fp4"

	// QuantizationFP6 represents Floating point (6 bit) quantization
	QuantizationFP6 Quantization = "fp6"

	// QuantizationFP8 represents Floating point (8 bit) quantization
	QuantizationFP8 Quantization = "fp8"

	// QuantizationFP16 represents Floating point (16 bit) quantization
	QuantizationFP16 Quantization = "fp16"

	// QuantizationBF16 represents Brain floating point (16 bit) quantization
	QuantizationBF16 Quantization = "bf16"

	// QuantizationFP32 represents Floating point (32 bit) quantization
	QuantizationFP32 Quantization = "fp32"

	// QuantizationUnknown represents Unknown quantization level
	QuantizationUnknown Quantization = "unknown"
)

// CompletionRequest represents a request payload for the completions endpoint.
// It contains all parameters needed to generate text completions from AI models.
type CompletionRequest struct {
	// Required fields
	// Model is the identifier of the AI model to use
	Model string `json:"model"`
	// Prompt is the text input that the model will complete
	Prompt string `json:"prompt"`

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
}

// UsageOptions controls whether to include token usage information in the response.
// When enabled, the API will return counts of prompt, completion, and total tokens.
type UsageOptions struct {
	// Include determines whether token usage information should be returned
	Include *bool `json:"usage,omitempty"`
}

// ReasoningOptions configures how models allocate tokens for internal reasoning.
// This allows models to "think" before producing a final response.
type ReasoningOptions struct {
	// Effort sets the proportion of tokens to allocate for reasoning
	Effort Effort `json:"effort,omitempty"`
	// MaxTokens sets the maximum number of tokens for reasoning
	MaxTokens *int `json:"max_tokens,omitempty"`
	// Exclude determines whether to include reasoning in the final response
	Exclude *bool `json:"exclude,omitempty"`
}

// CompletionChoice represents a single completion result from the API.
// The API may return multiple choices depending on the request parameters.
type CompletionChoice struct {
	// Text is the generated completion content
	Text string `json:"text"`
	// Index is the position of this choice in the array of choices
	Index int `json:"index"`
	// FinishReason explains why the generation stopped (e.g., "length", "stop")
	FinishReason string `json:"finish_reason"`
}

// Usage provides detailed information about token consumption for a request.
// This helps with monitoring costs and optimizing prompts.
type Usage struct {
	// PromptTokens is the number of tokens in the input prompt
	PromptTokens int `json:"prompt_tokens"`
	// CompletionTokens is the number of tokens in the generated completion
	CompletionTokens int `json:"completion_tokens"`
	// TotalTokens is the sum of prompt and completion tokens
	TotalTokens int `json:"total_tokens"`
}

// CompletionRequestBuilder implements a builder pattern for constructing CompletionRequest objects.
// This makes it easier to create requests with many optional parameters.
type CompletionRequestBuilder struct {
	request CompletionRequest
}

// NewCompletionRequestBuilder creates a new builder initialized with the required model and prompt.
//
// Parameters:
//   - model: The identifier of the AI model to use
//   - prompt: The text prompt that the model will complete
//
// Returns:
//   - *CompletionRequestBuilder: A builder instance that can be used to set optional parameters
func NewCompletionRequestBuilder(model, prompt string) *CompletionRequestBuilder {
	return &CompletionRequestBuilder{
		request: CompletionRequest{
			Model:  model,
			Prompt: prompt,
		},
	}
}

// WithModels sets the list of alternative models
func (b *CompletionRequestBuilder) WithModels(models []string) *CompletionRequestBuilder {
	b.request.Models = models
	return b
}

// WithProvider sets provider routing options
func (b *CompletionRequestBuilder) WithProvider(provider *ProviderOptions) *CompletionRequestBuilder {
	b.request.Provider = provider
	return b
}

// WithReasoning sets reasoning options
func (b *CompletionRequestBuilder) WithReasoning(reasoning *ReasoningOptions) *CompletionRequestBuilder {
	b.request.Reasoning = reasoning
	return b
}

// WithUsage sets usage information option
func (b *CompletionRequestBuilder) WithUsage(usage bool) *CompletionRequestBuilder {
	if b.request.Usage == nil {
		b.request.Usage = &UsageOptions{}
	}
	b.request.Usage.Include = &usage
	return b
}

// WithTransforms sets prompt transforms
func (b *CompletionRequestBuilder) WithTransforms(transforms []string) *CompletionRequestBuilder {
	b.request.Transforms = transforms
	return b
}

// WithStream enables or disables streaming
func (b *CompletionRequestBuilder) WithStream(stream bool) *CompletionRequestBuilder {
	b.request.Stream = &stream
	return b
}

// WithMaxTokens sets the maximum tokens
func (b *CompletionRequestBuilder) WithMaxTokens(maxTokens int) *CompletionRequestBuilder {
	b.request.MaxTokens = &maxTokens
	return b
}

// WithTemperature sets the sampling temperature
func (b *CompletionRequestBuilder) WithTemperature(temperature float64) *CompletionRequestBuilder {
	b.request.Temperature = &temperature
	return b
}

// WithSeed sets the seed for deterministic outputs
func (b *CompletionRequestBuilder) WithSeed(seed int) *CompletionRequestBuilder {
	b.request.Seed = &seed
	return b
}

// WithTopP sets the top-p sampling value
func (b *CompletionRequestBuilder) WithTopP(topP float64) *CompletionRequestBuilder {
	b.request.TopP = &topP
	return b
}

// WithTopK sets the top-k sampling value
func (b *CompletionRequestBuilder) WithTopK(topK int) *CompletionRequestBuilder {
	b.request.TopK = &topK
	return b
}

// WithFrequencyPenalty sets the frequency penalty
func (b *CompletionRequestBuilder) WithFrequencyPenalty(penalty float64) *CompletionRequestBuilder {
	b.request.FrequencyPenalty = &penalty
	return b
}

// WithPresencePenalty sets the presence penalty
func (b *CompletionRequestBuilder) WithPresencePenalty(penalty float64) *CompletionRequestBuilder {
	b.request.PresencePenalty = &penalty
	return b
}

// WithRepetitionPenalty sets the repetition penalty
func (b *CompletionRequestBuilder) WithRepetitionPenalty(penalty float64) *CompletionRequestBuilder {
	b.request.RepetitionPenalty = &penalty
	return b
}

// WithLogitBias sets the logit bias map
func (b *CompletionRequestBuilder) WithLogitBias(biases map[string]float64) *CompletionRequestBuilder {
	b.request.LogitBias = biases
	return b
}

// WithTopLogprobs sets the number of top log probabilities to return
func (b *CompletionRequestBuilder) WithTopLogprobs(topLogProbs int) *CompletionRequestBuilder {
	b.request.TopLogProbs = &topLogProbs
	return b
}

// WithMinP sets the minimum probability threshold
func (b *CompletionRequestBuilder) WithMinP(minP float64) *CompletionRequestBuilder {
	b.request.MinP = &minP
	return b
}

// WithTopA sets the alternate top sampling parameter
func (b *CompletionRequestBuilder) WithTopA(topA float64) *CompletionRequestBuilder {
	b.request.TopA = &topA
	return b
}

// WithLogprobs enables or disables returning log probabilities of output tokens
func (b *CompletionRequestBuilder) WithLogprobs(logprobs bool) *CompletionRequestBuilder {
	b.request.Logprobs = &logprobs
	return b
}

// WithStop sets the stop sequences for token generation
func (b *CompletionRequestBuilder) WithStop(stop []string) *CompletionRequestBuilder {
	b.request.Stop = stop
	return b
}

// Build finalizes and returns the constructed CompletionRequest.
//
// Returns:
//   - CompletionRequest: The fully constructed request object with all configured parameters
func (b *CompletionRequestBuilder) Build() CompletionRequest {
	return b.request
}

// ProviderOptions specifies preferences for how OpenRouter should route requests to AI providers.
// These options allow for fine-grained control over which providers are used and how they are selected.
type ProviderOptions struct {
	// AllowFallbacks determines whether to try backup providers when the primary is unavailable
	AllowFallbacks *bool `json:"allow_fallbacks,omitempty"`

	// RequireParameters ensures only providers that support all request parameters are used
	RequireParameters *bool `json:"require_parameters,omitempty"`

	// DataCollection controls whether to use providers that may store data
	// Valid values: "deny", "allow"
	DataCollection string `json:"data_collection,omitempty"`

	// Order specifies the ordered list of provider names to try (e.g. ["Anthropic", "OpenAI"])
	Order []string `json:"order,omitempty"`

	// Only limits request routing to only the specified providers
	Only []string `json:"only,omitempty"`

	// Ignore specifies which providers should not be used for this request
	Ignore []string `json:"ignore,omitempty"`

	// Quantizations filters providers by their model quantization levels
	// Valid values include: "int4", "int8", "fp4", "fp6", "fp8", "fp16", "bf16", "fp32", "unknown"
	Quantizations []Quantization `json:"quantizations,omitempty"`

	// Sort specifies how to rank available providers
	// Valid values: "price", "throughput", "latency"
	Sort string `json:"sort,omitempty"`

	// MaxPrice sets the maximum pricing limits for this request
	MaxPrice *MaxPrice `json:"max_price,omitempty"`

	// Experimental contains experimental provider routing features
	Experimental *ExperimentalOptions `json:"experimental,omitempty"`
}

// MaxPrice specifies the maximum price limits for different components of a request.
// All prices are in USD and allow for cost control when using the API.
type MaxPrice struct {
	// Prompt is the maximum USD price per million tokens for the input prompt
	Prompt *float64 `json:"prompt,omitempty"`

	// Completion is the maximum USD price per million tokens for the generated completion
	Completion *float64 `json:"completion,omitempty"`

	// Image is the maximum USD price per image included in the request
	Image *float64 `json:"image,omitempty"`

	// Request is the maximum USD price per API request regardless of tokens
	Request *float64 `json:"request,omitempty"`
}

// ExperimentalOptions contains cutting-edge features that may change in future API versions.
// These options provide additional control for advanced use cases.
type ExperimentalOptions struct {
	// ForceChatCompletions forces the use of chat completions API even when using the completions endpoint
	ForceChatCompletions *bool `json:"force_chat_completions,omitempty"`
}

// ProviderOptionsBuilder implements a builder pattern for constructing ProviderOptions objects.
// This provides a fluent interface for configuring the many options available for provider routing.
type ProviderOptionsBuilder struct {
	options ProviderOptions
}

// NewProviderOptionsBuilder creates a new builder for configuring provider routing options.
// The returned builder can be used to set options through method chaining.
func NewProviderOptionsBuilder() *ProviderOptionsBuilder {
	return &ProviderOptionsBuilder{}
}

// WithAllowFallbacks sets whether to allow backup providers
func (b *ProviderOptionsBuilder) WithAllowFallbacks(allow bool) *ProviderOptionsBuilder {
	b.options.AllowFallbacks = &allow
	return b
}

// WithRequireParameters sets whether to require providers to support all parameters
func (b *ProviderOptionsBuilder) WithRequireParameters(require bool) *ProviderOptionsBuilder {
	b.options.RequireParameters = &require
	return b
}

// WithDataCollection sets the data collection policy
// Values should be "allow" or "deny"
func (b *ProviderOptionsBuilder) WithDataCollection(policy string) *ProviderOptionsBuilder {
	b.options.DataCollection = policy
	return b
}

// WithOrder sets the list of provider names to try in order
func (b *ProviderOptionsBuilder) WithOrder(providers []string) *ProviderOptionsBuilder {
	b.options.Order = providers
	return b
}

// WithOnly sets the list of provider names to exclusively allow
func (b *ProviderOptionsBuilder) WithOnly(providers []string) *ProviderOptionsBuilder {
	b.options.Only = providers
	return b
}

// WithIgnore sets the list of provider names to skip
func (b *ProviderOptionsBuilder) WithIgnore(providers []string) *ProviderOptionsBuilder {
	b.options.Ignore = providers
	return b
}

// WithQuantizations sets the list of quantization levels to filter by
func (b *ProviderOptionsBuilder) WithQuantizations(quantizations []Quantization) *ProviderOptionsBuilder {
	b.options.Quantizations = quantizations
	return b
}

// WithSort sets the sorting strategy
// Values should be "price", "throughput", or "latency"
func (b *ProviderOptionsBuilder) WithSort(sort string) *ProviderOptionsBuilder {
	b.options.Sort = sort
	return b
}

// WithMaxPrice sets the maximum pricing configuration
func (b *ProviderOptionsBuilder) WithMaxPrice(maxPrice *MaxPrice) *ProviderOptionsBuilder {
	b.options.MaxPrice = maxPrice
	return b
}

// WithMaxPromptPrice sets the maximum price per million prompt tokens
func (b *ProviderOptionsBuilder) WithMaxPromptPrice(price float64) *ProviderOptionsBuilder {
	if b.options.MaxPrice == nil {
		b.options.MaxPrice = &MaxPrice{}
	}
	b.options.MaxPrice.Prompt = &price
	return b
}

// WithMaxCompletionPrice sets the maximum price per million completion tokens
func (b *ProviderOptionsBuilder) WithMaxCompletionPrice(price float64) *ProviderOptionsBuilder {
	if b.options.MaxPrice == nil {
		b.options.MaxPrice = &MaxPrice{}
	}
	b.options.MaxPrice.Completion = &price
	return b
}

// WithMaxImagePrice sets the maximum price per image
func (b *ProviderOptionsBuilder) WithMaxImagePrice(price float64) *ProviderOptionsBuilder {
	if b.options.MaxPrice == nil {
		b.options.MaxPrice = &MaxPrice{}
	}
	b.options.MaxPrice.Image = &price
	return b
}

// WithMaxRequestPrice sets the maximum price per request
func (b *ProviderOptionsBuilder) WithMaxRequestPrice(price float64) *ProviderOptionsBuilder {
	if b.options.MaxPrice == nil {
		b.options.MaxPrice = &MaxPrice{}
	}
	b.options.MaxPrice.Request = &price
	return b
}

// WithForceChatCompletions sets whether to force using chat completions API
func (b *ProviderOptionsBuilder) WithForceChatCompletions(force bool) *ProviderOptionsBuilder {
	if b.options.Experimental == nil {
		b.options.Experimental = &ExperimentalOptions{}
	}
	b.options.Experimental.ForceChatCompletions = &force
	return b
}

// Build finalizes and returns the constructed ProviderOptions.
//
// Returns:
//   - *ProviderOptions: A pointer to the fully configured provider options object
func (b *ProviderOptionsBuilder) Build() *ProviderOptions {
	return &b.options
}

// CompletionResponse represents the API response from a text completion request.
// It contains the generated completions and associated metadata.
type CompletionResponse struct {
	// ID is the unique identifier for this completion request
	ID string `json:"id"`
	// Model is the name of the model that generated the completion
	Model string `json:"model"`
	// Choices contains the generated text completions
	Choices []CompletionChoice `json:"choices"`
	// Usage provides token usage statistics for the request
	Usage Usage `json:"usage"`
}

// Completion sends a text completion request to the OpenRouter API.
//
// This method allows users to generate text completions from AI models through the
// OpenRouter API. The request can be customized with various parameters to control
// the generation process, provider selection, and output format.
//
// Parameters:
//   - ctx: The context for the request, which can be used for cancellation and timeouts
//   - request: The completion request parameters
//
// Returns:
//   - CompletionResponse: Contains the generated completions and metadata
//   - error: Any error that occurred during the request, including ErrCompletionStreamNotSupported
//     if streaming was requested
func (c *Client) Completion(
	ctx context.Context,
	request CompletionRequest,
) (response CompletionResponse, err error) {
	if request.Stream != nil && *request.Stream {
		err = ErrCompletionStreamNotSupported
		return
	}

	urlSuffix := "/completions"

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
