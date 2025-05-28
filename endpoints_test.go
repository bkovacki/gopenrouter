package gopenrouter_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/bkovacki/gopenrouter"
)

func TestListEndpoints(t *testing.T) {
	cases := []struct {
		name         string
		author       string
		slug         string
		handler      http.HandlerFunc
		expectErr    bool
		expectAPIErr bool
		expectReqErr bool
		validateResp func(t *testing.T, data gopenrouter.EndpointData)
	}{
		{
			name:   "Success",
			author: "test-author",
			slug:   "test-model",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				expectedPath := "/models/test-author/test-model/endpoints"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{
					"data": {
						"id": "test-author/test-model",
						"name": "Test Model",
						"created": 1622505600,
						"description": "A test model",
						"architecture": {
							"input_modalities": ["text"],
							"output_modalities": ["text"],
							"tokenizer": "test-tokenizer",
							"instruct_type": "test-instruct"
						},
						"endpoints": [
							{
								"name": "provider-1-endpoint",
								"context_length": 4096,
								"pricing": {
									"request": "0.0001",
									"image": "0.0005",
									"prompt": "0.0002",
									"completion": "0.0003"
								},
								"provider_name": "Provider1",
								"supported_parameters": ["temperature", "max_tokens"]
							},
							{
								"name": "provider-2-endpoint",
								"context_length": 8192,
								"pricing": {
									"request": "0.0002",
									"image": "0.0006",
									"prompt": "0.0003",
									"completion": "0.0004"
								},
								"provider_name": "Provider2",
								"supported_parameters": ["temperature", "max_tokens", "top_p"]
							}
						]
					}
				}`)
			},
			expectErr: false,
			validateResp: func(t *testing.T, data gopenrouter.EndpointData) {
				// Validate model metadata
				if data.ID != "test-author/test-model" {
					t.Errorf("Expected ID 'test-author/test-model', got '%s'", data.ID)
				}
				if data.Name != "Test Model" {
					t.Errorf("Expected Name 'Test Model', got '%s'", data.Name)
				}
				if data.Created != 1622505600 {
					t.Errorf("Expected Created 1622505600, got %f", data.Created)
				}
				if data.Description != "A test model" {
					t.Errorf("Expected Description 'A test model', got '%s'", data.Description)
				}

				// Validate architecture
				expectedInputModalities := []string{"text"}
				if !reflect.DeepEqual(data.Architecture.InputModalities, expectedInputModalities) {
					t.Errorf("Expected InputModalities %v, got %v", expectedInputModalities, data.Architecture.InputModalities)
				}
				expectedOutputModalities := []string{"text"}
				if !reflect.DeepEqual(data.Architecture.OutputModalities, expectedOutputModalities) {
					t.Errorf("Expected OutputModalities %v, got %v", expectedOutputModalities, data.Architecture.OutputModalities)
				}
				if data.Architecture.Tokenizer != "test-tokenizer" {
					t.Errorf("Expected Tokenizer 'test-tokenizer', got '%s'", data.Architecture.Tokenizer)
				}
				if data.Architecture.InstructType != "test-instruct" {
					t.Errorf("Expected InstructType 'test-instruct', got '%s'", data.Architecture.InstructType)
				}

				// Validate endpoints
				if len(data.Endpoints) != 2 {
					t.Errorf("Expected 2 endpoints, got %d", len(data.Endpoints))
					return
				}

				// Validate first endpoint
				endpoint1 := data.Endpoints[0]
				if endpoint1.Name != "provider-1-endpoint" {
					t.Errorf("Expected endpoint1 Name 'provider-1-endpoint', got '%s'", endpoint1.Name)
				}
				if endpoint1.ContextLength != 4096 {
					t.Errorf("Expected endpoint1 ContextLength 4096, got %f", endpoint1.ContextLength)
				}
				if endpoint1.ProviderName != "Provider1" {
					t.Errorf("Expected endpoint1 ProviderName 'Provider1', got '%s'", endpoint1.ProviderName)
				}
				expectedParams1 := []string{"temperature", "max_tokens"}
				if !reflect.DeepEqual(endpoint1.SupportedParameters, expectedParams1) {
					t.Errorf("Expected endpoint1 SupportedParameters %v, got %v", expectedParams1, endpoint1.SupportedParameters)
				}
				if endpoint1.Pricing.Request != "0.0001" {
					t.Errorf("Expected endpoint1 Pricing.Request '0.0001', got '%s'", endpoint1.Pricing.Request)
				}
				if endpoint1.Pricing.Image != "0.0005" {
					t.Errorf("Expected endpoint1 Pricing.Image '0.0005', got '%s'", endpoint1.Pricing.Image)
				}
				if endpoint1.Pricing.Prompt != "0.0002" {
					t.Errorf("Expected endpoint1 Pricing.Prompt '0.0002', got '%s'", endpoint1.Pricing.Prompt)
				}
				if endpoint1.Pricing.Completion != "0.0003" {
					t.Errorf("Expected endpoint1 Pricing.Completion '0.0003', got '%s'", endpoint1.Pricing.Completion)
				}

				// Validate second endpoint
				endpoint2 := data.Endpoints[1]
				if endpoint2.Name != "provider-2-endpoint" {
					t.Errorf("Expected endpoint2 Name 'provider-2-endpoint', got '%s'", endpoint2.Name)
				}
				if endpoint2.ContextLength != 8192 {
					t.Errorf("Expected endpoint2 ContextLength 8192, got %f", endpoint2.ContextLength)
				}
				if endpoint2.ProviderName != "Provider2" {
					t.Errorf("Expected endpoint2 ProviderName 'Provider2', got '%s'", endpoint2.ProviderName)
				}
				expectedParams2 := []string{"temperature", "max_tokens", "top_p"}
				if !reflect.DeepEqual(endpoint2.SupportedParameters, expectedParams2) {
					t.Errorf("Expected endpoint2 SupportedParameters %v, got %v", expectedParams2, endpoint2.SupportedParameters)
				}
			},
		},
		{
			name:   "APIError",
			author: "test-author",
			slug:   "test-model",
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
			name:   "UnexpectedHTMLResponse",
			author: "test-author",
			slug:   "test-model",
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
			name:   "SpecialCharactersInPath",
			author: "special-author@with/chars",
			slug:   "special-model#name",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Check if the path is correctly URL-encoded with special characters
				expectedPath := "/models/special-author@with%2Fchars/special-model%23name/endpoints"

				// Use the raw path (which preserves URL encoding) or request URI for checking
				actualPath := r.URL.RawPath
				if actualPath == "" {
					actualPath = r.RequestURI
				}

				if actualPath != expectedPath {
					t.Errorf("Expected encoded path %s, got %s", expectedPath, actualPath)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{"data": {"id": "test-id", "name": "Test Model"}}`)
			},
			expectErr: false,
			validateResp: func(t *testing.T, data gopenrouter.EndpointData) {
				if data.ID != "test-id" {
					t.Errorf("Expected ID 'test-id', got '%s'", data.ID)
				}
			},
		},
		{
			name:   "URLEncodingTest",
			author: "openai",
			slug:   "gpt-4-turbo",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Just verify we can access this endpoint with normal model names
				expectedPath := "/models/openai/gpt-4-turbo/endpoints"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{"data": {"id": "openai/gpt-4-turbo", "name": "GPT-4 Turbo"}}`)
			},
			expectErr: false,
			validateResp: func(t *testing.T, data gopenrouter.EndpointData) {
				if data.ID != "openai/gpt-4-turbo" {
					t.Errorf("Expected ID 'openai/gpt-4-turbo', got '%s'", data.ID)
				}
				if data.Name != "GPT-4 Turbo" {
					t.Errorf("Expected Name 'GPT-4 Turbo', got '%s'", data.Name)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(tc.handler)
			defer ts.Close()

			client := gopenrouter.New("test-key", gopenrouter.WithBaseURL(ts.URL))
			data, err := client.ListEndpoints(context.Background(), tc.author, tc.slug)

			var apiErr *gopenrouter.APIError
			var reqErr *gopenrouter.RequestError

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
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
				if tc.validateResp != nil {
					tc.validateResp(t, data)
				}
			}
		})
	}
}
