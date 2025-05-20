package gopenrouter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewClientDefaults(t *testing.T) {
	apiKey := "test-api-key"
	client := New(apiKey)

	if client.apiKey != apiKey {
		t.Errorf("expected apiKey %q, got %q", apiKey, client.apiKey)
	}
	if client.baseURL != openRouterAPIURL {
		t.Errorf("expected baseURL %q, got %q", openRouterAPIURL, client.baseURL)
	}
	if client.httpClient != http.DefaultClient {
		t.Error("expected httpClient to be http.DefaultClient")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	apiKey := "test-api-key"
	siteURL := "https://testing.com"
	siteTitle := "GOpenRouter"
	baseURL := "https://test.ai"
	httpClient := &http.Client{}

	client := New(apiKey,
		WithSiteURL(siteURL),
		WithSiteTitle(siteTitle),
		WithBaseURL(baseURL),
		WithHTTPClient(httpClient))

	if client.siteURL != siteURL {
		t.Errorf("expected siteURL %q, got %q", siteURL, client.siteURL)
	}
	if client.siteTitle != siteTitle {
		t.Errorf("expected siteTitle %q, got %q", siteTitle, client.siteTitle)
	}
	if client.baseURL != baseURL {
		t.Errorf("expected baseURL %q, got %q", baseURL, client.baseURL)
	}
	if client.httpClient != httpClient {
		t.Errorf("expected httpClient %p, got %p", httpClient, client.httpClient)
	}
}

func TestClientSetCommonHeaders(t *testing.T) {
	apiKey := "test-api-key"
	siteURL := "https://testing.com"
	siteTitle := "GOpenRouter"

	client := New(apiKey, WithSiteURL(siteURL), WithSiteTitle(siteTitle))
	req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
	client.setCommonHeaders(req)

	if req.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", apiKey) {
		t.Error("Authorization header not set")
	}
	if req.Header.Get("HTTP-Referer") != siteURL {
		t.Error("HTTP-Referer header not set")
	}
	if req.Header.Get("X-Title") != siteTitle {
		t.Error("X-Title header not set")
	}
}

func TestHandleErrorResp(t *testing.T) {
	cases := []struct {
		name         string
		body         string
		statusCode   int
		expectAPIErr bool
		expectCode   int
		expectMsg    string
	}{
		{
			name:         "valid ErrorResponse",
			body:         `{"error": {"code": 400, "message": "Invalid request"}}`,
			statusCode:   400,
			expectAPIErr: true,
			expectCode:   400,
			expectMsg:    "Invalid request",
		},
		{
			name:         "invalid JSON returns RequestError",
			body:         `not a json`,
			statusCode:   500,
			expectAPIErr: false,
			expectCode:   500,
			expectMsg:    "",
		},
	}

	apiKey := "test-api-key"
	client := New(apiKey)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tc.statusCode,
				Body:       io.NopCloser(strings.NewReader(tc.body)),
				Header:     make(http.Header),
			}

			err := client.handleErrorResp(resp)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var apiErr *APIError
			var reqErr *RequestError
			if tc.expectAPIErr {
				if errors.As(err, &apiErr) {
					errStr := apiErr.Error()
					if apiErr.Code != tc.expectCode || apiErr.Message != tc.expectMsg {
						t.Errorf("unexpected APIError: %+v", apiErr)
					}
					if apiErr.Code > 0 && !strings.Contains(errStr, fmt.Sprintf("%d", apiErr.Code)) {
						t.Errorf("Error() string does not contain code: %s", errStr)
					}
					if !strings.Contains(errStr, apiErr.Message) {
						t.Errorf("Error() string does not contain message: %s", errStr)
					}
				} else {
					t.Errorf("expected APIError, got %T: %v", err, err)
				}
			} else {
				if errors.As(err, &reqErr) {
					errStr := reqErr.Error()
					if reqErr.HTTPStatusCode != tc.expectCode {
						t.Errorf("unexpected status code in RequestError: %+v", reqErr)
					}
					if !strings.Contains(errStr, fmt.Sprintf("%d", reqErr.HTTPStatusCode)) {
						t.Errorf("Error() string does not contain status code: %s", errStr)
					}
					if reqErr.Err != nil && !strings.Contains(errStr, reqErr.Err.Error()) {
						t.Errorf("Error() string does not contain wrapped error: %s", errStr)
					}
				} else {
					t.Errorf("expected RequestError, got %T: %v", err, err)
				}
			}
		})
	}
}

func TestClientNewRequest(t *testing.T) {
	client := New("test-key", WithBaseURL("https://api.example.com"))

	type payload struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name            string
		method          string
		path            string
		setters         []requestOption
		wantURL         string
		wantBody        string
		wantContentType string
	}{
		{
			name:            "POST with body and query params",
			method:          http.MethodPost,
			path:            "/users",
			setters:         []requestOption{withQueryParam("foo", "bar"), withBody(payload{Name: "test"})},
			wantURL:         "https://api.example.com/users?foo=bar",
			wantBody:        `{"name":"test"}`,
			wantContentType: "application/json",
		},
		{
			name:            "GET with query params",
			method:          http.MethodGet,
			path:            "/users",
			setters:         []requestOption{withQueryParam("a", "b")},
			wantURL:         "https://api.example.com/users?a=b",
			wantContentType: "application/json",
		},
		{
			name:            "GET without query params",
			method:          http.MethodGet,
			path:            "/users",
			wantURL:         "https://api.example.com/users",
			wantContentType: "application/json",
		},
		{
			name:            "POST with custom content type",
			method:          http.MethodPost,
			path:            "/custom",
			setters:         []requestOption{withContentType("application/x-custom"), withBody(payload{Name: "custom"})},
			wantURL:         "https://api.example.com/custom",
			wantBody:        `{"name":"custom"}`,
			wantContentType: "application/x-custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := client.newRequest(context.Background(), tt.method, client.fullURL(tt.path), tt.setters...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if req.Method != tt.method {
				t.Errorf("expected method %s, got %s", tt.method, req.Method)
			}
			if req.URL.String() != tt.wantURL {
				t.Errorf("expected URL %s, got %s", tt.wantURL, req.URL.String())
			}
			if tt.wantBody != "" {
				b, _ := io.ReadAll(req.Body)
				if string(b) != tt.wantBody {
					t.Errorf("expected body %s, got %s", tt.wantBody, string(b))
				}
			}
			if ct := req.Header.Get("Content-Type"); ct != tt.wantContentType {
				t.Errorf("expected Content-Type %q, got %q", tt.wantContentType, ct)
			}
		})
	}
}
