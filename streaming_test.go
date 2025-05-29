package gopenrouter

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCompletionStream(t *testing.T) {
	t.Run("SuccessfulStream", func(t *testing.T) {
		// Mock server that sends streaming response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.WriteHeader(http.StatusOK)

			// Send streaming chunks
			chunks := []string{
				`data: {"id":"gen-1748550593-SiBpqgpnEC1joxVF6DZZ","provider":"OpenAI","model":"openai/gpt-4o-mini","object":"chat.completion.chunk","created":1748550593,"choices":[{"index":0,"text":"Hello","finish_reason":null,"native_finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_34a54ae93c"}`,
				`data: {"id":"gen-1748550593-SiBpqgpnEC1joxVF6DZZ","provider":"OpenAI","model":"openai/gpt-4o-mini","object":"chat.completion.chunk","created":1748550593,"choices":[{"index":0,"text":" world","finish_reason":null,"native_finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_34a54ae93c"}`,
				`data: {"id":"gen-1748550593-SiBpqgpnEC1joxVF6DZZ","provider":"OpenAI","model":"openai/gpt-4o-mini","object":"chat.completion.chunk","created":1748550593,"choices":[{"index":0,"text":"!","finish_reason":"stop","native_finish_reason":"stop","logprobs":null}],"system_fingerprint":"fp_34a54ae93c"}`,
				`data: {"id":"gen-1748550593-SiBpqgpnEC1joxVF6DZZ","provider":"OpenAI","model":"openai/gpt-4o-mini","object":"chat.completion.chunk","created":1748550593,"choices":[{"index":0,"text":"","finish_reason":null,"native_finish_reason":null,"logprobs":null}],"usage":{"prompt_tokens":16,"completion_tokens":61,"total_tokens":77}}`,
				`data: [DONE]`,
			}

			for _, chunk := range chunks {
				w.Write([]byte(chunk + "\n\n"))
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
		}))
		defer server.Close()

		client := New("test-api-key", WithBaseURL(server.URL))
		request := NewCompletionRequestBuilder("test-model", "test prompt").Build()

		stream, err := client.CompletionStream(context.Background(), request)
		if err != nil {
			t.Fatalf("CompletionStream failed: %v", err)
		}
		defer stream.Close()

		// Read first chunk
		chunk1, err := stream.Recv()
		if err != nil {
			t.Fatalf("Failed to read first chunk: %v", err)
		}
		if chunk1.ID != "gen-1748550593-SiBpqgpnEC1joxVF6DZZ" {
			t.Errorf("Expected ID 'gen-1748550593-SiBpqgpnEC1joxVF6DZZ', got '%s'", chunk1.ID)
		}
		if chunk1.Provider != "OpenAI" {
			t.Errorf("Expected provider 'OpenAI', got '%s'", chunk1.Provider)
		}
		if chunk1.Model != "openai/gpt-4o-mini" {
			t.Errorf("Expected model 'openai/gpt-4o-mini', got '%s'", chunk1.Model)
		}
		if chunk1.SystemFingerprint == nil || *chunk1.SystemFingerprint != "fp_34a54ae93c" {
			t.Errorf("Expected system_fingerprint 'fp_34a54ae93c', got %v", chunk1.SystemFingerprint)
		}
		if len(chunk1.Choices) != 1 {
			t.Errorf("Expected 1 choice, got %d", len(chunk1.Choices))
		}
		if chunk1.Choices[0].Text != "Hello" {
			t.Errorf("Expected text 'Hello', got '%s'", chunk1.Choices[0].Text)
		}

		// Read second chunk
		_, err = stream.Recv()
		if err != nil {
			t.Fatalf("Failed to read second chunk: %v", err)
		}

		// Read third chunk
		chunk3, err := stream.Recv()
		if err != nil {
			t.Fatalf("Failed to read third chunk: %v", err)
		}
		if chunk3.Choices[0].FinishReason == nil || *chunk3.Choices[0].FinishReason != "stop" {
			t.Errorf("Expected finish_reason 'stop', got %v", chunk3.Choices[0].FinishReason)
		}
		if chunk3.Choices[0].NativeFinishReason == nil || *chunk3.Choices[0].NativeFinishReason != "stop" {
			t.Errorf("Expected native_finish_reason 'stop', got %v", chunk3.Choices[0].NativeFinishReason)
		}

		// Read fourth chunk (usage data)
		chunk4, err := stream.Recv()
		if err != nil {
			t.Fatalf("Failed to read fourth chunk: %v", err)
		}
		if chunk4.Usage == nil {
			t.Error("Expected usage data in final chunk")
		} else {
			if chunk4.Usage.PromptTokens != 16 {
				t.Errorf("Expected prompt_tokens 16, got %d", chunk4.Usage.PromptTokens)
			}
			if chunk4.Usage.CompletionTokens != 61 {
				t.Errorf("Expected completion_tokens 61, got %d", chunk4.Usage.CompletionTokens)
			}
			if chunk4.Usage.TotalTokens != 77 {
				t.Errorf("Expected total_tokens 77, got %d", chunk4.Usage.TotalTokens)
			}
		}

		// Read final chunk - should return EOF
		_, err = stream.Recv()
		if err != io.EOF {
			t.Errorf("Expected EOF at end of stream, got %v", err)
		}
	})

	t.Run("StreamWithComments", func(t *testing.T) {
		// Mock server that sends comments (OpenRouter processing indicators)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)

			chunks := []string{
				`: OPENROUTER PROCESSING`,
				`data: {"id":"gen-1748550593-SiBpqgpnEC1joxVF6DZZ","provider":"OpenAI","model":"openai/gpt-4o-mini","object":"chat.completion.chunk","created":1748550593,"choices":[{"index":0,"text":"Hello","finish_reason":null,"native_finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_34a54ae93c"}`,
				`: Keep-alive comment`,
				`data: [DONE]`,
			}

			for _, chunk := range chunks {
				w.Write([]byte(chunk + "\n"))
			}
		}))
		defer server.Close()

		client := New("test-api-key", WithBaseURL(server.URL))
		request := NewCompletionRequestBuilder("test-model", "test prompt").Build()

		stream, err := client.CompletionStream(context.Background(), request)
		if err != nil {
			t.Fatalf("CompletionStream failed: %v", err)
		}
		defer stream.Close()

		// Should skip comments and return the data chunk
		chunk, err := stream.Recv()
		if err != nil {
			t.Fatalf("Failed to read chunk: %v", err)
		}
		if chunk.ID != "gen-1748550593-SiBpqgpnEC1joxVF6DZZ" {
			t.Errorf("Expected ID 'gen-1748550593-SiBpqgpnEC1joxVF6DZZ', got '%s'", chunk.ID)
		}
		if chunk.Provider != "OpenAI" {
			t.Errorf("Expected provider 'OpenAI', got '%s'", chunk.Provider)
		}
		if chunk.SystemFingerprint == nil || *chunk.SystemFingerprint != "fp_34a54ae93c" {
			t.Errorf("Expected system_fingerprint 'fp_34a54ae93c', got %v", chunk.SystemFingerprint)
		}

		// Next should be EOF
		_, err = stream.Recv()
		if err != io.EOF {
			t.Errorf("Expected EOF, got %v", err)
		}
	})

	t.Run("EmptyResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("data: [DONE]\n"))
		}))
		defer server.Close()

		client := New("test-api-key", WithBaseURL(server.URL))
		request := NewCompletionRequestBuilder("test-model", "test prompt").Build()

		stream, err := client.CompletionStream(context.Background(), request)
		if err != nil {
			t.Fatalf("CompletionStream failed: %v", err)
		}
		defer stream.Close()

		_, err = stream.Recv()
		if err != io.EOF {
			t.Errorf("Expected EOF for empty stream, got %v", err)
		}
	})

	t.Run("ServerError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":{"message":"Internal server error"}}`))
		}))
		defer server.Close()

		client := New("test-api-key", WithBaseURL(server.URL))
		request := NewCompletionRequestBuilder("test-model", "test prompt").Build()

		_, err := client.CompletionStream(context.Background(), request)
		if err == nil {
			t.Error("Expected error for server error response")
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
				`data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":1234567890,"model":"test-model","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}]}`,
				`data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":1234567890,"model":"test-model","choices":[{"index":0,"delta":{"content":" there"},"finish_reason":null}]}`,
				`data: {"id":"chatcmpl-1","object":"chat.completion.chunk","created":1234567890,"model":"test-model","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":"stop"}]}`,
				`data: [DONE]`,
			}

			for _, chunk := range chunks {
				w.Write([]byte(chunk + "\n\n"))
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
		}))
		defer server.Close()

		client := New("test-api-key", WithBaseURL(server.URL))
		messages := []ChatMessage{{Role: "user", Content: "Hello"}}
		request := NewChatCompletionRequestBuilder("test-model", messages).Build()

		stream, err := client.ChatCompletionStream(context.Background(), *request)
		if err != nil {
			t.Fatalf("ChatCompletionStream failed: %v", err)
		}
		defer stream.Close()

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

		// Read second chunk
		chunk2, err := stream.Recv()
		if err != nil {
			t.Fatalf("Failed to read second chunk: %v", err)
		}
		if chunk2.Choices[0].Delta.Content == nil || *chunk2.Choices[0].Delta.Content != " there" {
			t.Errorf("Expected content ' there' in second chunk")
		}

		// Read third chunk
		chunk3, err := stream.Recv()
		if err != nil {
			t.Fatalf("Failed to read third chunk: %v", err)
		}
		if chunk3.Choices[0].FinishReason == nil || *chunk3.Choices[0].FinishReason != "stop" {
			t.Errorf("Expected finish_reason 'stop', got %v", chunk3.Choices[0].FinishReason)
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
			w.Write([]byte("data: [DONE]\n"))
		}))
		defer server.Close()

		client := New("test-api-key", WithBaseURL(server.URL))
		messages := []ChatMessage{{Role: "user", Content: "Hello"}}
		
		// Create request without explicitly setting stream=true
		request := NewChatCompletionRequestBuilder("test-model", messages).Build()
		
		stream, err := client.ChatCompletionStream(context.Background(), *request)
		if err != nil {
			t.Fatalf("ChatCompletionStream failed: %v", err)
		}
		defer stream.Close()

		// Stream should be handled internally - we don't modify the original request
		// Just verify that the streaming endpoint was called successfully
	})


}

func TestStreamReaderClose(t *testing.T) {
	t.Run("CompletionStreamReaderClose", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			// Don't send [DONE] to test early close
			w.Write([]byte(`data: {"id":"gen-1748550593-SiBpqgpnEC1joxVF6DZZ","provider":"OpenAI","model":"openai/gpt-4o-mini","object":"chat.completion.chunk","created":1748550593,"choices":[{"index":0,"text":"test","finish_reason":null,"native_finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_34a54ae93c"}` + "\n"))
		}))
		defer server.Close()

		client := New("test-api-key", WithBaseURL(server.URL))
		request := NewCompletionRequestBuilder("test-model", "test").Build()

		stream, err := client.CompletionStream(context.Background(), request)
		if err != nil {
			t.Fatalf("CompletionStream failed: %v", err)
		}

		// Close immediately
		err = stream.Close()
		if err != nil {
			t.Errorf("Close failed: %v", err)
		}

		// Subsequent reads should fail
		_, err = stream.Recv()
		if err == nil {
			t.Error("Expected error after closing stream")
		}
	})

	t.Run("ChatCompletionStreamReaderClose", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`data: {"id":"chat-1","choices":[{"index":0,"delta":{"content":"test"}}]}` + "\n"))
		}))
		defer server.Close()

		client := New("test-api-key", WithBaseURL(server.URL))
		messages := []ChatMessage{{Role: "user", Content: "Hello"}}
		request := NewChatCompletionRequestBuilder("test-model", messages).Build()

		stream, err := client.ChatCompletionStream(context.Background(), *request)
		if err != nil {
			t.Fatalf("ChatCompletionStream failed: %v", err)
		}

		err = stream.Close()
		if err != nil {
			t.Errorf("Close failed: %v", err)
		}
	})
}

func TestStreamContextCancellation(t *testing.T) {
	t.Run("ContextCancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			
			// Send one chunk then delay
			w.Write([]byte(`data: {"id":"gen-1748550593-SiBpqgpnEC1joxVF6DZZ","provider":"OpenAI","model":"openai/gpt-4o-mini","object":"chat.completion.chunk","created":1748550593,"choices":[{"index":0,"text":"test","finish_reason":null,"native_finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_34a54ae93c"}` + "\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			
			// Long delay to allow context cancellation
			time.Sleep(100 * time.Millisecond)
			w.Write([]byte("data: [DONE]\n"))
		}))
		defer server.Close()

		client := New("test-api-key", WithBaseURL(server.URL))
		request := NewCompletionRequestBuilder("test-model", "test").Build()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		stream, err := client.CompletionStream(ctx, request)
		if err != nil {
			// Context might be cancelled before stream is created
			if strings.Contains(err.Error(), "context deadline exceeded") {
				return // This is acceptable
			}
			t.Fatalf("CompletionStream failed: %v", err)
		}
		defer stream.Close()

		// Read first chunk should work
		_, err = stream.Recv()
		if err != nil && !strings.Contains(err.Error(), "context deadline exceeded") {
			t.Fatalf("Unexpected error: %v", err)
		}
	})
}

func TestMalformedStreamData(t *testing.T) {
	t.Run("InvalidJSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)

			chunks := []string{
				`data: {invalid json}`,
				`data: {"id":"gen-1748550593-SiBpqgpnEC1joxVF6DZZ","provider":"OpenAI","model":"openai/gpt-4o-mini","object":"chat.completion.chunk","created":1748550593,"choices":[{"index":0,"text":"valid","finish_reason":null,"native_finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_34a54ae93c"}`,
				`data: [DONE]`,
			}

			for _, chunk := range chunks {
				w.Write([]byte(chunk + "\n"))
			}
		}))
		defer server.Close()

		client := New("test-api-key", WithBaseURL(server.URL))
		request := NewCompletionRequestBuilder("test-model", "test").Build()

		stream, err := client.CompletionStream(context.Background(), request)
		if err != nil {
			t.Fatalf("CompletionStream failed: %v", err)
		}
		defer stream.Close()

		// Should skip invalid JSON and return valid chunk
		chunk, err := stream.Recv()
		if err != nil {
			t.Fatalf("Failed to read valid chunk: %v", err)
		}
		if chunk.ID != "gen-1748550593-SiBpqgpnEC1joxVF6DZZ" {
			t.Errorf("Expected valid chunk with ID 'gen-1748550593-SiBpqgpnEC1joxVF6DZZ', got '%s'", chunk.ID)
		}
		if chunk.Provider != "OpenAI" {
			t.Errorf("Expected provider 'OpenAI', got '%s'", chunk.Provider)
		}

		// Next should be EOF
		_, err = stream.Recv()
		if err != io.EOF {
			t.Errorf("Expected EOF, got %v", err)
		}
	})
}