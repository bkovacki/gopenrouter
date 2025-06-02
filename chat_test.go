package gopenrouter_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bkovacki/gopenrouter"
)

func TestChatCompletionRequestBuilder(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		messages := []gopenrouter.ChatMessage{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "What is the capital of France?"},
		}

		builder := gopenrouter.NewChatCompletionRequestBuilder("openai/gpt-4", messages)
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
		messages := []gopenrouter.ChatMessage{
			{Role: "user", Content: "Test message"},
		}

		providerOptions := gopenrouter.NewProviderOptionsBuilder().
			WithAllowFallbacks(true).
			WithMaxPromptPrice(0.01).
			Build()

		reasoningOptions := &gopenrouter.ReasoningOptions{
			Effort:    gopenrouter.EffortMedium,
			MaxTokens: &[]int{50}[0],
			Exclude:   &[]bool{false}[0],
		}

		builder := gopenrouter.NewChatCompletionRequestBuilder("openai/gpt-3.5-turbo", messages)
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
			if request.Reasoning.Effort != gopenrouter.EffortMedium {
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
		messages := []gopenrouter.ChatMessage{
			{Role: "user", Content: "Test sampling parameters"},
		}

		logitBias := map[string]float64{
			"1000": -100,
			"2000": 100,
		}

		builder := gopenrouter.NewChatCompletionRequestBuilder("openai/gpt-3.5-turbo", messages)
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
			WithLogprobs(true).
			WithStop([]string{"STOP", "END"}).
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

		if request.Logprobs == nil || *request.Logprobs != true {
			t.Errorf("Expected logprobs to be true, got %v", request.Logprobs)
		}

		if len(request.Stop) != 2 || request.Stop[0] != "STOP" || request.Stop[1] != "END" {
			t.Errorf("Expected stop to be [STOP, END], got %v", request.Stop)
		}
	})
}

func TestChatCompletion(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Mock response
		mockResponse := gopenrouter.ChatCompletionResponse{
			ID: "gen-12345",
			Choices: []gopenrouter.ChatChoice{
				{
					Message: gopenrouter.ChatMessage{
						Role:    "assistant",
						Content: "The capital of France is Paris.",
					},
					Index:        0,
					FinishReason: "stop",
				},
			},
			Usage: gopenrouter.Usage{
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
			var req gopenrouter.ChatCompletionRequest
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
		client := gopenrouter.New("test-api-key", gopenrouter.WithBaseURL(server.URL))

		// Create request
		messages := []gopenrouter.ChatMessage{
			{Role: "user", Content: "What is the capital of France?"},
		}

		request := gopenrouter.NewChatCompletionRequestBuilder("openai/gpt-3.5-turbo", messages).Build()

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
		client := gopenrouter.New("test-api-key")

		messages := []gopenrouter.ChatMessage{
			{Role: "user", Content: "Test message"},
		}

		request := gopenrouter.NewChatCompletionRequestBuilder("openai/gpt-3.5-turbo", messages).
			WithStream(true).
			Build()

		ctx := context.Background()
		_, err := client.ChatCompletion(ctx, *request)

		if err != gopenrouter.ErrCompletionStreamNotSupported {
			t.Errorf("Expected ErrCompletionStreamNotSupported, got %v", err)
		}
	})
}

func TestChatCompletionStream(t *testing.T) {
	t.Run("SuccessfulStream", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusOK)

			chunks := []string{
				`data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":1234567890,"model":"test-model","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null,"logprobs":{"content":[{"token":"Hello","bytes":[72,101,108,108,111],"logprob":-0.8,"top_logprobs":[{"token":"Hello","bytes":[72,101,108,108,111],"logprob":-0.8},{"token":"Hi","bytes":[72,105],"logprob":-1.5}]}],"refusal":[]}}]}`,
				`data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":1234567890,"model":"test-model","choices":[{"index":0,"delta":{"content":" there"},"finish_reason":null,"logprobs":{"content":[{"token":" there","bytes":[32,116,104,101,114,101],"logprob":-0.2,"top_logprobs":[{"token":" there","bytes":[32,116,104,101,114,101],"logprob":-0.2},{"token":" world","bytes":[32,119,111,114,108,100],"logprob":-2.1}]}],"refusal":[]}}]}`,
				`data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":1234567890,"model":"test-model","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":"stop","logprobs":{"content":[{"token":"!","bytes":[33],"logprob":-0.1,"top_logprobs":[{"token":"!","bytes":[33],"logprob":-0.1},{"token":".","bytes":[46],"logprob":-2.8}]}],"refusal":[]}}]}`,
				`data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":1234567890,"model":"test-model","choices":[{"index":0,"delta":{},"finish_reason":null,"logprobs":null}],"usage":{"prompt_tokens":5,"completion_tokens":3,"total_tokens":8,"prompt_tokens_details":{"cached_tokens":1},"completion_tokens_details":{"reasoning_tokens":0}}}`,
				`data: [DONE]`,
			}

			for _, chunk := range chunks {
				_, _ = w.Write([]byte(chunk + "\n\n"))
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
		}))
		defer server.Close()

		client := gopenrouter.New("test-api-key", gopenrouter.WithBaseURL(server.URL))
		messages := []gopenrouter.ChatMessage{{Role: "user", Content: "Hello"}}
		request := gopenrouter.NewChatCompletionRequestBuilder("test-model", messages).Build()

		stream, err := client.ChatCompletionStream(context.Background(), *request)
		if err != nil {
			t.Fatalf("ChatCompletionStream failed: %v", err)
		}
		defer func() { _ = stream.Close() }()

		// Read first chunk
		chunk1, err := stream.Recv()
		if err != nil {
			t.Fatalf("Failed to read first chunk: %v", err)
		}
		if chunk1.ID != "chatcmpl-1" {
			t.Errorf("Expected ID 'chatcmpl-1', got '%s'", chunk1.ID)
		}
		if len(chunk1.Choices) != 1 {
			t.Errorf("Expected 1 choice, got %d", len(chunk1.Choices))
		}
		if chunk1.Choices[0].Delta.Role == nil || *chunk1.Choices[0].Delta.Role != "assistant" {
			t.Errorf("Expected role 'assistant' in first chunk")
		}
		if chunk1.Choices[0].Delta.Content == nil || *chunk1.Choices[0].Delta.Content != "Hello" {
			t.Errorf("Expected content 'Hello' in first chunk")
		}

		// Verify logprobs in first chunk
		if chunk1.Choices[0].LogProbs == nil {
			t.Error("Expected LogProbs to be non-nil in first chunk")
		} else {
			if len(chunk1.Choices[0].LogProbs.Content) != 1 {
				t.Errorf("Expected 1 content token in first chunk, got %d", len(chunk1.Choices[0].LogProbs.Content))
			} else {
				token := chunk1.Choices[0].LogProbs.Content[0]
				if token.Token != "Hello" {
					t.Errorf("Expected token 'Hello' in first chunk, got '%s'", token.Token)
				}
				if token.LogProb != -0.8 {
					t.Errorf("Expected logprob -0.8 in first chunk, got %f", token.LogProb)
				}
				if len(token.TopLogProbs) != 2 {
					t.Errorf("Expected 2 top logprobs in first chunk, got %d", len(token.TopLogProbs))
				}
			}
		}

		// Read second chunk
		chunk2, err := stream.Recv()
		if err != nil {
			t.Fatalf("Failed to read second chunk: %v", err)
		}
		if chunk2.Choices[0].Delta.Content == nil || *chunk2.Choices[0].Delta.Content != " there" {
			t.Errorf("Expected content ' there' in second chunk")
		}

		// Verify logprobs in second chunk
		if chunk2.Choices[0].LogProbs == nil {
			t.Error("Expected LogProbs to be non-nil in second chunk")
		} else {
			if len(chunk2.Choices[0].LogProbs.Content) != 1 {
				t.Errorf("Expected 1 content token in second chunk, got %d", len(chunk2.Choices[0].LogProbs.Content))
			} else {
				token := chunk2.Choices[0].LogProbs.Content[0]
				if token.Token != " there" {
					t.Errorf("Expected token ' there' in second chunk, got '%s'", token.Token)
				}
				if token.LogProb != -0.2 {
					t.Errorf("Expected logprob -0.2 in second chunk, got %f", token.LogProb)
				}
			}
		}

		// Read third chunk
		chunk3, err := stream.Recv()
		if err != nil {
			t.Fatalf("Failed to read third chunk: %v", err)
		}
		if chunk3.Choices[0].FinishReason == nil || *chunk3.Choices[0].FinishReason != "stop" {
			t.Errorf("Expected finish_reason 'stop', got %v", chunk3.Choices[0].FinishReason)
		}

		// Verify logprobs in third chunk
		if chunk3.Choices[0].LogProbs == nil {
			t.Error("Expected LogProbs to be non-nil in third chunk")
		} else {
			if len(chunk3.Choices[0].LogProbs.Content) != 1 {
				t.Errorf("Expected 1 content token in third chunk, got %d", len(chunk3.Choices[0].LogProbs.Content))
			} else {
				token := chunk3.Choices[0].LogProbs.Content[0]
				if token.Token != "!" {
					t.Errorf("Expected token '!' in third chunk, got '%s'", token.Token)
				}
				if token.LogProb != -0.1 {
					t.Errorf("Expected logprob -0.1 in third chunk, got %f", token.LogProb)
				}
			}
		}

		// Read usage chunk
		chunk4, err := stream.Recv()
		if err != nil {
			t.Fatalf("Failed to read usage chunk: %v", err)
		}
		if chunk4.Usage == nil {
			t.Error("Expected Usage to be non-nil in usage chunk")
		} else {
			if chunk4.Usage.TotalTokens != 8 {
				t.Errorf("Expected total tokens 8, got %d", chunk4.Usage.TotalTokens)
			}
			if chunk4.Usage.PromptTokensDetails == nil {
				t.Error("Expected PromptTokensDetails to be non-nil")
			} else if chunk4.Usage.PromptTokensDetails.CachedTokens != 1 {
				t.Errorf("Expected cached tokens 1, got %d", chunk4.Usage.PromptTokensDetails.CachedTokens)
			}
			if chunk4.Usage.CompletionTokensDetails == nil {
				t.Error("Expected CompletionTokensDetails to be non-nil")
			} else if chunk4.Usage.CompletionTokensDetails.ReasoningTokens != 0 {
				t.Errorf("Expected reasoning tokens 0, got %d", chunk4.Usage.CompletionTokensDetails.ReasoningTokens)
			}
		}

		// Read final chunk - should return EOF
		_, err = stream.Recv()
		if err != io.EOF {
			t.Errorf("Expected EOF at end of stream, got %v", err)
		}
	})

	t.Run("StreamAutomaticallyEnabled", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify that stream parameter was set to true
			if r.Header.Get("Accept") != "text/event-stream" {
				t.Errorf("Expected Accept header 'text/event-stream'")
			}

			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("data: [DONE]\n"))
		}))
		defer server.Close()

		client := gopenrouter.New("test-api-key", gopenrouter.WithBaseURL(server.URL))
		messages := []gopenrouter.ChatMessage{{Role: "user", Content: "Hello"}}

		// Create request without explicitly setting stream=true
		request := gopenrouter.NewChatCompletionRequestBuilder("test-model", messages).Build()

		stream, err := client.ChatCompletionStream(context.Background(), *request)
		if err != nil {
			t.Fatalf("ChatCompletionStream failed: %v", err)
		}
		defer func() { _ = stream.Close() }()

		// Stream should be handled internally - we don't modify the original request
		// Just verify that the streaming endpoint was called successfully
	})

}

func TestChatStreamReaderClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`data: {"id":"chat-1","choices":[{"index":0,"delta":{"content":"test"}}]}` + "\n"))
	}))
	defer server.Close()

	client := gopenrouter.New("test-api-key", gopenrouter.WithBaseURL(server.URL))
	messages := []gopenrouter.ChatMessage{{Role: "user", Content: "Hello"}}
	request := gopenrouter.NewChatCompletionRequestBuilder("test-model", messages).Build()

	stream, err := client.ChatCompletionStream(context.Background(), *request)
	if err != nil {
		t.Fatalf("ChatCompletionStream failed: %v", err)
	}

	err = stream.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
