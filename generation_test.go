package gopenrouter_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bkovacki/gopenrouter"
)

func TestClientGetGeneration(t *testing.T) {
	cases := []struct {
		name         string
		handler      http.HandlerFunc
		expectErr    bool
		expectAPIErr bool
		expectReqErr bool
		expectID     string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{
					"data": {
						"id": "gen-123",
						"total_cost": 1.1,
						"created_at": "2024-01-01T00:00:00Z",
						"model": "test-model",
						"origin": "origin",
						"usage": 1.1,
						"is_byok": true,
						"upstream_id": "upstream_id",
						"cache_discount": 1.1,
						"app_id": 1,
						"streamed": true,
						"cancelled": false,
						"provider_name": "provider",
						"latency": 10,
						"moderation_latency": 2,
						"generation_time": 5,
						"finish_reason": "stop",
						"native_finish_reason": "stop",
						"tokens_prompt": 5,
						"tokens_completion": 10,
						"native_tokens_prompt": 5,
						"native_tokens_completion": 10,
						"native_tokens_reasoning": 0,
						"num_media_prompt": 0,
						"num_media_completion": 0,
						"num_search_results": 0
					}
				}`)
			},
			expectErr: false,
			expectID:  "gen-123",
		},
		{
			name: "api error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{"error": {"code": 400, "message": "Invalid generation id"}}`)
			},
			expectErr:    true,
			expectAPIErr: true,
		},
		{
			name: "unexpected html",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "text/html")
				_, _ = fmt.Fprint(w, `<html><body>Internal Server Error</body></html>`)
			},
			expectErr:    true,
			expectReqErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(tc.handler)
			defer ts.Close()

			client := gopenrouter.New("test-key", gopenrouter.WithBaseURL(ts.URL))
			resp, err := client.GetGeneration(context.Background(), "gen-123")

			var apiErr *gopenrouter.APIError
			var reqErr *gopenrouter.RequestError

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.expectAPIErr && !errors.As(err, &apiErr) {
					t.Errorf("expected APIError, got %T: %v", err, err)
				}
				if tc.expectReqErr && !errors.As(err, &reqErr) {
					t.Errorf("expected RequestError, got %T: %v", err, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp.ID != tc.expectID {
					t.Errorf("unexpected generation ID: got %s, want %s", resp.ID, tc.expectID)
				}
			}
		})
	}
}
