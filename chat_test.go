package gopenrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChatCompletionRequestBuilder(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		messages := []ChatMessage{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "What is the capital of France?"},
		}

		builder := NewChatCompletionRequestBuilder("openai/gpt-4", messages)
		request := builder.
			WithMaxTokens(100).
			WithTemperature(0.7).
			WithStream(false).
			WithUser("test-user").
			Build()

		if request.Model != "openai/gpt-4" {
			t.Errorf("Expected model to be 'openai/gpt-4', got %s", request.Model)
		}

		if len(request.Messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(request.Messages))
		}

		if request.MaxTokens == nil || *request.MaxTokens != 100 {
			t.Errorf("Expected max_tokens to be 100, got %v", request.MaxTokens)
		}

		if request.Temperature == nil || *request.Temperature != 0.7 {
			t.Errorf("Expected temperature to be 0.7, got %v", request.Temperature)
		}

		if request.Stream == nil || *request.Stream != false {
			t.Errorf("Expected stream to be false, got %v", request.Stream)
		}

		if request.User == nil || *request.User != "test-user" {
			t.Errorf("Expected user to be 'test-user', got %v", request.User)
		}
	})

	t.Run("WithProviderOptions", func(t *testing.T) {
		messages := []ChatMessage{
			{Role: "user", Content: "Test message"},
		}

		providerOptions := NewProviderOptionsBuilder().
			WithAllowFallbacks(true).
			WithMaxPromptPrice(0.01).
			Build()

		reasoningOptions := &ReasoningOptions{
			Effort:    EffortMedium,
			MaxTokens: &[]int{50}[0],
			Exclude:   &[]bool{false}[0],
		}

		builder := NewChatCompletionRequestBuilder("openai/gpt-3.5-turbo", messages)
		request := builder.
			WithProvider(providerOptions).
			WithReasoning(reasoningOptions).
			WithUsage(true).
			WithTransforms([]string{"middle-out"}).
			Build()

		if request.Provider == nil {
			t.Error("Expected provider options to be set")
		}

		if request.Reasoning == nil {
			t.Error("Expected reasoning options to be set")
		} else {
			if request.Reasoning.Effort != EffortMedium {
				t.Errorf("Expected reasoning effort to be medium, got %s", request.Reasoning.Effort)
			}
		}

		if request.Usage == nil || request.Usage.Include == nil || !*request.Usage.Include {
			t.Error("Expected usage to be enabled")
		}

		if len(request.Transforms) != 1 || request.Transforms[0] != "middle-out" {
			t.Errorf("Expected transforms to contain 'middle-out', got %v", request.Transforms)
		}
	})

	t.Run("WithSamplingParameters", func(t *testing.T) {
		messages := []ChatMessage{
			{Role: "user", Content: "Test sampling parameters"},
		}

		logitBias := map[string]float64{
			"1000": -100,
			"2000": 100,
		}

		builder := NewChatCompletionRequestBuilder("openai/gpt-3.5-turbo", messages)
		request := builder.
			WithTopP(0.9).
			WithTopK(50).
			WithFrequencyPenalty(0.5).
			WithPresencePenalty(0.3).
			WithRepetitionPenalty(1.1).
			WithLogitBias(logitBias).
			WithTopLogprobs(5).
			WithMinP(0.05).
			WithTopA(0.2).
			WithSeed(42).
			Build()

		if request.TopP == nil || *request.TopP != 0.9 {
			t.Errorf("Expected top_p to be 0.9, got %v", request.TopP)
		}

		if request.TopK == nil || *request.TopK != 50 {
			t.Errorf("Expected top_k to be 50, got %v", request.TopK)
		}

		if request.FrequencyPenalty == nil || *request.FrequencyPenalty != 0.5 {
			t.Errorf("Expected frequency_penalty to be 0.5, got %v", request.FrequencyPenalty)
		}

		if request.PresencePenalty == nil || *request.PresencePenalty != 0.3 {
			t.Errorf("Expected presence_penalty to be 0.3, got %v", request.PresencePenalty)
		}

		if request.RepetitionPenalty == nil || *request.RepetitionPenalty != 1.1 {
			t.Errorf("Expected repetition_penalty to be 1.1, got %v", request.RepetitionPenalty)
		}

		if len(request.LogitBias) != 2 {
			t.Errorf("Expected 2 logit bias entries, got %d", len(request.LogitBias))
		}

		if request.LogitBias["1000"] != -100 {
			t.Errorf("Expected logit bias for token 1000 to be -100, got %f", request.LogitBias["1000"])
		}

		if request.TopLogProbs == nil || *request.TopLogProbs != 5 {
			t.Errorf("Expected top_logprobs to be 5, got %v", request.TopLogProbs)
		}

		if request.MinP == nil || *request.MinP != 0.05 {
			t.Errorf("Expected min_p to be 0.05, got %v", request.MinP)
		}

		if request.TopA == nil || *request.TopA != 0.2 {
			t.Errorf("Expected top_a to be 0.2, got %v", request.TopA)
		}

		if request.Seed == nil || *request.Seed != 42 {
			t.Errorf("Expected seed to be 42, got %v", request.Seed)
		}
	})
}

func TestChatCompletion(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Mock response
		mockResponse := ChatCompletionResponse{
			ID: "gen-12345",
			Choices: []ChatChoice{
				{
					Message: ChatMessage{
						Role:    "assistant",
						Content: "The capital of France is Paris.",
					},
					Index:        0,
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 8,
				TotalTokens:      18,
			},
		}

		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			if r.URL.Path != "/chat/completions" {
				t.Errorf("Expected path /chat/completions, got %s", r.URL.Path)
			}

			// Check authorization header
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				t.Error("Expected Authorization header with Bearer token")
			}

			// Check content type
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}

			// Parse request body
			var req ChatCompletionRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			// Validate request
			if req.Model != "openai/gpt-3.5-turbo" {
				t.Errorf("Expected model openai/gpt-3.5-turbo, got %s", req.Model)
			}

			if len(req.Messages) != 1 {
				t.Errorf("Expected 1 message, got %d", len(req.Messages))
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(mockResponse); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
		}))
		defer server.Close()

		// Create client
		client := New("test-api-key", WithBaseURL(server.URL))

		// Create request
		messages := []ChatMessage{
			{Role: "user", Content: "What is the capital of France?"},
		}

		request := NewChatCompletionRequestBuilder("openai/gpt-3.5-turbo", messages).Build()

		// Make request
		ctx := context.Background()
		response, err := client.ChatCompletion(ctx, *request)
		if err != nil {
			t.Fatalf("ChatCompletion failed: %v", err)
		}

		// Validate response
		if response.ID != "gen-12345" {
			t.Errorf("Expected ID gen-12345, got %s", response.ID)
		}

		if len(response.Choices) != 1 {
			t.Errorf("Expected 1 choice, got %d", len(response.Choices))
		}

		choice := response.Choices[0]
		if choice.Message.Role != "assistant" {
			t.Errorf("Expected role assistant, got %s", choice.Message.Role)
		}

		if choice.Message.Content != "The capital of France is Paris." {
			t.Errorf("Expected specific content, got %s", choice.Message.Content)
		}

		if choice.FinishReason != "stop" {
			t.Errorf("Expected finish reason stop, got %s", choice.FinishReason)
		}

		if response.Usage.TotalTokens != 18 {
			t.Errorf("Expected total tokens 18, got %d", response.Usage.TotalTokens)
		}
	})

	t.Run("StreamNotSupported", func(t *testing.T) {
		client := New("test-api-key")

		messages := []ChatMessage{
			{Role: "user", Content: "Test message"},
		}

		request := NewChatCompletionRequestBuilder("openai/gpt-3.5-turbo", messages).
			WithStream(true).
			Build()

		ctx := context.Background()
		_, err := client.ChatCompletion(ctx, *request)

		if err != ErrCompletionStreamNotSupported {
			t.Errorf("Expected ErrCompletionStreamNotSupported, got %v", err)
		}
	})
}
