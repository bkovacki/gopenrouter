package gopenrouter_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bkovacki/gopenrouter"
)

func TestClientCompletion(t *testing.T) {
	cases := []struct {
		name           string
		handler        http.HandlerFunc
		request        gopenrouter.CompletionRequest
		expectErr      bool
		expectAPIErr   bool
		expectReqErr   bool
		expectErrType  error // For specific error types like ErrCompletionStreamNotSupported
		expectRespText string
	}{
		{
			name: "Success",
			request: gopenrouter.NewCompletionRequestBuilder(
				"test-model",
				"Say hello",
			).Build(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{
                    "id": "cmpl-123",
                    "model": "test-model",
                    "choices": [
                        {"text": "Hello, world!", "index": 0, "finish_reason": "stop"}
                    ],
                    "usage": {"prompt_tokens": 1, "completion_tokens": 2, "total_tokens": 3}
                }`)
			},
			expectErr:      false,
			expectAPIErr:   false,
			expectReqErr:   false,
			expectRespText: "Hello, world!",
		},
		{
			name: "APIError",
			request: gopenrouter.NewCompletionRequestBuilder(
				"test-model",
				"Say hello",
			).Build(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{"error": {"code": 400, "message": "Invalid request"}}`)
			},
			expectErr:    true,
			expectAPIErr: true,
			expectReqErr: false,
		},
		{
			name: "UnexpectedHTML",
			request: gopenrouter.NewCompletionRequestBuilder(
				"test-model",
				"Say hello",
			).Build(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "text/html")
				_, _ = fmt.Fprint(w, `<html><body>Internal Server Error</body></html>`)
			},
			expectErr:    true,
			expectAPIErr: false,
			expectReqErr: true,
		},
		{
			name: "StreamNotSupported",
			request: gopenrouter.NewCompletionRequestBuilder(
				"test-model",
				"Say hello",
			).WithStream(true).Build(),
			handler:       nil, // No handler needed as error occurs before HTTP request
			expectErr:     true,
			expectErrType: gopenrouter.ErrCompletionStreamNotSupported,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(tc.handler)
			defer ts.Close()

			client := gopenrouter.New("test-key", gopenrouter.WithBaseURL(ts.URL))

			resp, err := client.Completion(context.Background(), tc.request)
			var apiErr *gopenrouter.APIError
			var reqErr *gopenrouter.RequestError

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.expectErrType != nil {
					if !errors.Is(err, tc.expectErrType) {
						t.Errorf("expected error type %v, got %v", tc.expectErrType, err)
					}
				}
				if tc.expectAPIErr && !errors.As(err, &apiErr) {
					t.Errorf("expected APIError, got %T: %v", err, err)
				}
				if !tc.expectAPIErr && errors.As(err, &apiErr) {
					t.Errorf("did not expect APIError, got one: %v", apiErr)
				}
				if tc.expectReqErr && !errors.As(err, &reqErr) {
					t.Errorf("expected RequestError, got %T: %v", err, err)
				}
				if !tc.expectReqErr && errors.As(err, &reqErr) {
					t.Errorf("did not expect RequestError, got one: %v", reqErr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(resp.Choices) == 0 || resp.Choices[0].Text != tc.expectRespText {
					t.Errorf("unexpected choices: %+v", resp.Choices)
				}
			}
		})
	}
}

func TestCompletionRequestBuilder(t *testing.T) {
	testModel := "test-model"
	testPrompt := "test-prompt"

	t.Run("ConstructionWithRequiredFields", func(t *testing.T) {

		builder := gopenrouter.NewCompletionRequestBuilder(testModel, testPrompt)
		request := builder.Build()

		if request.Model != testModel {
			t.Errorf("Expected model to be %s, got %q", testModel, request.Model)
		}
		if request.Prompt != testPrompt {
			t.Errorf("Expected prompt to be %s, got %q", testPrompt, request.Prompt)
		}
	})

	t.Run("WithAllScalarOptions", func(t *testing.T) {
		stream := true
		maxTokens := 100
		temperature := 0.7
		seed := 42
		topP := 0.9
		topK := 5
		frequencyPenalty := 0.5
		presencePenalty := 0.3
		repetitionPenalty := 1.2
		topLogProbs := 3
		minP := 0.1
		topA := 0.8
		logprobs := true
		stop := []string{"stop1", "stop2"}

		builder := gopenrouter.NewCompletionRequestBuilder(testModel, testPrompt)
		request := builder.
			WithStream(stream).
			WithMaxTokens(maxTokens).
			WithTemperature(temperature).
			WithSeed(seed).
			WithTopP(topP).
			WithTopK(topK).
			WithFrequencyPenalty(frequencyPenalty).
			WithPresencePenalty(presencePenalty).
			WithRepetitionPenalty(repetitionPenalty).
			WithTopLogprobs(topLogProbs).
			WithMinP(minP).
			WithTopA(topA).
			WithLogprobs(logprobs).
			WithStop(stop).
			Build()

		if *request.Stream != stream {
			t.Errorf("Expected Stream to be %v, got %v", stream, *request.Stream)
		}
		if *request.MaxTokens != maxTokens {
			t.Errorf("Expected MaxTokens to be %d, got %d", maxTokens, *request.MaxTokens)
		}
		if *request.Temperature != temperature {
			t.Errorf("Expected Temperature to be %f, got %f", temperature, *request.Temperature)
		}
		if *request.Seed != seed {
			t.Errorf("Expected Seed to be %d, got %d", seed, *request.Seed)
		}
		if *request.TopP != topP {
			t.Errorf("Expected TopP to be %f, got %f", topP, *request.TopP)
		}
		if *request.TopK != topK {
			t.Errorf("Expected TopK to be %d, got %d", topK, *request.TopK)
		}
		if *request.FrequencyPenalty != frequencyPenalty {
			t.Errorf("Expected FrequencyPenalty to be %f, got %f", frequencyPenalty, *request.FrequencyPenalty)
		}
		if *request.PresencePenalty != presencePenalty {
			t.Errorf("Expected PresencePenalty to be %f, got %f", presencePenalty, *request.PresencePenalty)
		}
		if *request.RepetitionPenalty != repetitionPenalty {
			t.Errorf("Expected RepetitionPenalty to be %f, got %f", repetitionPenalty, *request.RepetitionPenalty)
		}
		if *request.TopLogProbs != topLogProbs {
			t.Errorf("Expected TopLogProbs to be %d, got %d", topLogProbs, *request.TopLogProbs)
		}
		if *request.MinP != minP {
			t.Errorf("Expected MinP to be %f, got %f", minP, *request.MinP)
		}
		if *request.TopA != topA {
			t.Errorf("Expected TopA to be %f, got %f", topA, *request.TopA)
		}
		if *request.Logprobs != logprobs {
			t.Errorf("Expected Logprobs to be %v, got %v", logprobs, *request.Logprobs)
		}
		if !reflect.DeepEqual(request.Stop, stop) {
			t.Errorf("Expected Stop to be %v, got %v", stop, request.Stop)
		}
	})

	t.Run("WithArrayAndMapOptions", func(t *testing.T) {
		models := []string{"model1", "model2"}
		transforms := []string{"transform1", "transform2"}
		logitBias := map[string]float64{"123": 1.0, "456": -1.0}
		stop := []string{"END", "STOP"}

		builder := gopenrouter.NewCompletionRequestBuilder(testModel, testPrompt)
		request := builder.
			WithModels(models).
			WithTransforms(transforms).
			WithLogitBias(logitBias).
			WithStop(stop).
			Build()

		if !reflect.DeepEqual(request.Models, models) {
			t.Errorf("Expected Models to be %v, got %v", models, request.Models)
		}
		if !reflect.DeepEqual(request.Transforms, transforms) {
			t.Errorf("Expected Transforms to be %v, got %v", transforms, request.Transforms)
		}
		if !reflect.DeepEqual(request.LogitBias, logitBias) {
			t.Errorf("Expected LogitBias to be %v, got %v", logitBias, request.LogitBias)
		}
		if !reflect.DeepEqual(request.Stop, stop) {
			t.Errorf("Expected Stop to be %v, got %v", stop, request.Stop)
		}
	})

	t.Run("WithUsageOption", func(t *testing.T) {
		builder := gopenrouter.NewCompletionRequestBuilder(testModel, testPrompt)
		request := builder.
			WithUsage(true).
			Build()

		if request.Usage == nil {
			t.Fatal("Expected Usage to be non-nil")
		}
		if *request.Usage.Include != true {
			t.Errorf("Expected Usage.Include to be true, got %v", *request.Usage.Include)
		}
	})

	t.Run("WithReasoningOption", func(t *testing.T) {
		maxTokens := 50
		exclude := true
		reasoning := &gopenrouter.ReasoningOptions{
			Effort:    gopenrouter.EffortHigh,
			MaxTokens: &maxTokens,
			Exclude:   &exclude,
		}

		builder := gopenrouter.NewCompletionRequestBuilder(testModel, testPrompt)
		request := builder.
			WithReasoning(reasoning).
			Build()

		if request.Reasoning == nil {
			t.Fatal("Expected Reasoning to be non-nil")
		}
		if request.Reasoning.Effort != gopenrouter.EffortHigh {
			t.Errorf("Expected Reasoning.Effort to be %v, got %v", gopenrouter.EffortHigh, request.Reasoning.Effort)
		}
		if *request.Reasoning.MaxTokens != maxTokens {
			t.Errorf("Expected Reasoning.MaxTokens to be %d, got %d", maxTokens, *request.Reasoning.MaxTokens)
		}
		if *request.Reasoning.Exclude != exclude {
			t.Errorf("Expected Reasoning.Exclude to be %v, got %v", exclude, *request.Reasoning.Exclude)
		}
	})

	t.Run("WithProviderOption", func(t *testing.T) {
		providerBuilder := gopenrouter.NewProviderOptionsBuilder()
		provider := providerBuilder.
			WithSort("price").
			Build()

		builder := gopenrouter.NewCompletionRequestBuilder(testModel, testPrompt)
		request := builder.
			WithProvider(provider).
			Build()

		if request.Provider == nil {
			t.Fatal("Expected Provider to be non-nil")
		}
		if request.Provider.Sort != "price" {
			t.Errorf("Expected Provider.Sort to be 'price', got %q", request.Provider.Sort)
		}
	})
}

func TestProviderOptionsBuilder(t *testing.T) {
	t.Run("EmptyBuilder", func(t *testing.T) {
		builder := gopenrouter.NewProviderOptionsBuilder()
		options := builder.Build()

		if options.AllowFallbacks != nil {
			t.Errorf("Expected AllowFallbacks to be nil, got %v", *options.AllowFallbacks)
		}
		if options.RequireParameters != nil {
			t.Errorf("Expected RequireParameters to be nil, got %v", *options.RequireParameters)
		}
		if options.DataCollection != "" {
			t.Errorf("Expected DataCollection to be empty, got %q", options.DataCollection)
		}
		if options.Sort != "" {
			t.Errorf("Expected Sort to be empty, got %q", options.Sort)
		}
	})

	t.Run("BooleanOptions", func(t *testing.T) {
		allowFallbacks := true
		requireParams := false
		forceChatCompletions := true

		builder := gopenrouter.NewProviderOptionsBuilder()
		options := builder.
			WithAllowFallbacks(allowFallbacks).
			WithRequireParameters(requireParams).
			WithForceChatCompletions(forceChatCompletions).
			Build()

		if *options.AllowFallbacks != allowFallbacks {
			t.Errorf("Expected AllowFallbacks to be %v, got %v", allowFallbacks, *options.AllowFallbacks)
		}
		if *options.RequireParameters != requireParams {
			t.Errorf("Expected RequireParameters to be %v, got %v", requireParams, *options.RequireParameters)
		}
		if options.Experimental == nil {
			t.Fatal("Expected Experimental to be non-nil")
		}
		if *options.Experimental.ForceChatCompletions != forceChatCompletions {
			t.Errorf("Expected ForceChatCompletions to be %v, got %v", forceChatCompletions, *options.Experimental.ForceChatCompletions)
		}
	})

	t.Run("StringOptions", func(t *testing.T) {
		dataCollection := "deny"
		sort := "latency"

		builder := gopenrouter.NewProviderOptionsBuilder()
		options := builder.
			WithDataCollection(dataCollection).
			WithSort(sort).
			Build()

		if options.DataCollection != dataCollection {
			t.Errorf("Expected DataCollection to be %q, got %q", dataCollection, options.DataCollection)
		}
		if options.Sort != sort {
			t.Errorf("Expected Sort to be %q, got %q", sort, options.Sort)
		}
	})

	t.Run("ArrayOptions", func(t *testing.T) {
		order := []string{"Anthropic", "OpenAI"}
		only := []string{"Anthropic"}
		ignore := []string{"Claude"}
		quantizations := []gopenrouter.Quantization{gopenrouter.QuantizationInt8, gopenrouter.QuantizationFP16}

		builder := gopenrouter.NewProviderOptionsBuilder()
		options := builder.
			WithOrder(order).
			WithOnly(only).
			WithIgnore(ignore).
			WithQuantizations(quantizations).
			Build()

		if !reflect.DeepEqual(options.Order, order) {
			t.Errorf("Expected Order to be %v, got %v", order, options.Order)
		}
		if !reflect.DeepEqual(options.Only, only) {
			t.Errorf("Expected Only to be %v, got %v", only, options.Only)
		}
		if !reflect.DeepEqual(options.Ignore, ignore) {
			t.Errorf("Expected Ignore to be %v, got %v", ignore, options.Ignore)
		}
		if !reflect.DeepEqual(options.Quantizations, quantizations) {
			t.Errorf("Expected Quantizations to be %v, got %v", quantizations, options.Quantizations)
		}
	})

	t.Run("MaxPriceOptionsWithFullObject", func(t *testing.T) {
		promptPrice := 0.001
		completionPrice := 0.002
		imagePrice := 0.01
		requestPrice := 0.003

		maxPrice := &gopenrouter.MaxPrice{
			Prompt:     &promptPrice,
			Completion: &completionPrice,
			Image:      &imagePrice,
			Request:    &requestPrice,
		}

		builder := gopenrouter.NewProviderOptionsBuilder()
		options := builder.
			WithMaxPrice(maxPrice).
			Build()

		if options.MaxPrice == nil {
			t.Fatal("Expected MaxPrice to be non-nil")
		}
		if *options.MaxPrice.Prompt != promptPrice {
			t.Errorf("Expected MaxPrice.Prompt to be %f, got %f", promptPrice, *options.MaxPrice.Prompt)
		}
		if *options.MaxPrice.Completion != completionPrice {
			t.Errorf("Expected MaxPrice.Completion to be %f, got %f", completionPrice, *options.MaxPrice.Completion)
		}
		if *options.MaxPrice.Image != imagePrice {
			t.Errorf("Expected MaxPrice.Image to be %f, got %f", imagePrice, *options.MaxPrice.Image)
		}
		if *options.MaxPrice.Request != requestPrice {
			t.Errorf("Expected MaxPrice.Request to be %f, got %f", requestPrice, *options.MaxPrice.Request)
		}
	})

	t.Run("MaxPriceOptionsWithIndividualSetters", func(t *testing.T) {
		promptPrice := 0.001
		completionPrice := 0.002
		imagePrice := 0.01
		requestPrice := 0.003

		builder := gopenrouter.NewProviderOptionsBuilder()
		options := builder.
			WithMaxPromptPrice(promptPrice).
			WithMaxCompletionPrice(completionPrice).
			WithMaxImagePrice(imagePrice).
			WithMaxRequestPrice(requestPrice).
			Build()

		if options.MaxPrice == nil {
			t.Fatal("Expected MaxPrice to be non-nil")
		}
		if *options.MaxPrice.Prompt != promptPrice {
			t.Errorf("Expected MaxPrice.Prompt to be %f, got %f", promptPrice, *options.MaxPrice.Prompt)
		}
		if *options.MaxPrice.Completion != completionPrice {
			t.Errorf("Expected MaxPrice.Completion to be %f, got %f", completionPrice, *options.MaxPrice.Completion)
		}
		if *options.MaxPrice.Image != imagePrice {
			t.Errorf("Expected MaxPrice.Image to be %f, got %f", imagePrice, *options.MaxPrice.Image)
		}
		if *options.MaxPrice.Request != requestPrice {
			t.Errorf("Expected MaxPrice.Request to be %f, got %f", requestPrice, *options.MaxPrice.Request)
		}
	})

	t.Run("MethodChaining", func(t *testing.T) {
		allowFallbacks := true
		dataCollection := "deny"
		order := []string{"Anthropic", "OpenAI"}
		sort := "price"

		builder := gopenrouter.NewProviderOptionsBuilder()
		options := builder.
			WithAllowFallbacks(allowFallbacks).
			WithDataCollection(dataCollection).
			WithOrder(order).
			WithSort(sort).
			Build()

		if *options.AllowFallbacks != allowFallbacks {
			t.Errorf("Expected AllowFallbacks to be %v, got %v", allowFallbacks, *options.AllowFallbacks)
		}
		if options.DataCollection != dataCollection {
			t.Errorf("Expected DataCollection to be %q, got %q", dataCollection, options.DataCollection)
		}
		if !reflect.DeepEqual(options.Order, order) {
			t.Errorf("Expected Order to be %v, got %v", order, options.Order)
		}
		if options.Sort != sort {
			t.Errorf("Expected Sort to be %q, got %q", sort, options.Sort)
		}
	})
}

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

		client := gopenrouter.New("test-api-key", gopenrouter.WithBaseURL(server.URL))
		request := gopenrouter.NewCompletionRequestBuilder("test-model", "test prompt").Build()

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

		client := gopenrouter.New("test-api-key", gopenrouter.WithBaseURL(server.URL))
		request := gopenrouter.NewCompletionRequestBuilder("test-model", "test prompt").Build()

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

		client := gopenrouter.New("test-api-key", gopenrouter.WithBaseURL(server.URL))
		request := gopenrouter.NewCompletionRequestBuilder("test-model", "test prompt").Build()

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

		client := gopenrouter.New("test-api-key", gopenrouter.WithBaseURL(server.URL))
		request := gopenrouter.NewCompletionRequestBuilder("test-model", "test prompt").Build()

		_, err := client.CompletionStream(context.Background(), request)
		if err == nil {
			t.Error("Expected error for server error response")
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

		client := gopenrouter.New("test-api-key", gopenrouter.WithBaseURL(server.URL))
		request := gopenrouter.NewCompletionRequestBuilder("test-model", "test").Build()

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

		client := gopenrouter.New("test-api-key", gopenrouter.WithBaseURL(server.URL))
		request := gopenrouter.NewCompletionRequestBuilder("test-model", "test").Build()

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

func TestStreamReaderClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		// Don't send [DONE] to test early close
		w.Write([]byte(`data: {"id":"gen-1748550593-SiBpqgpnEC1joxVF6DZZ","provider":"OpenAI","model":"openai/gpt-4o-mini","object":"chat.completion.chunk","created":1748550593,"choices":[{"index":0,"text":"test","finish_reason":null,"native_finish_reason":null,"logprobs":null}],"system_fingerprint":"fp_34a54ae93c"}` + "\n"))
	}))
	defer server.Close()

	client := gopenrouter.New("test-api-key", gopenrouter.WithBaseURL(server.URL))
	request := gopenrouter.NewCompletionRequestBuilder("test-model", "test").Build()

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
}
