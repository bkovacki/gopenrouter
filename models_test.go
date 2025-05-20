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

func TestClient_ListModels(t *testing.T) {
	cases := []struct {
		name         string
		handler      http.HandlerFunc
		expectErr    bool
		expectAPIErr bool
		expectReqErr bool
		expectCount  int
		expectFirst  string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{"data":[{"id":"model-1","name":"Model One","created":123,"description":"desc","architecture":{"input_modalities":["text"],"output_modalities":["text"],"tokenizer":"tok"},"top_provider":{"is_moderated":false},"pricing":{"prompt":"0.01","completion":"0.02","image":"0.03","request":"0.04","input_cache_read":"0.05","input_cache_write":"0.06","web_search":"0.07","internal_reasoning":"0.08"}}]}`)
			},
			expectErr:   false,
			expectCount: 1,
			expectFirst: "model-1",
		},
		{
			name: "api error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{"error": {"code": 400, "message": "Invalid API key"}}`)
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
			data, err := client.ListModels(context.Background())

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
				if len(data) != tc.expectCount {
					t.Errorf("unexpected model count: got %d, want %d", len(data), tc.expectCount)
				}
				if tc.expectCount > 0 && data[0].ID != tc.expectFirst {
					t.Errorf("unexpected first model ID: got %s, want %s", data[0].ID, tc.expectFirst)
				}
			}
		})
	}
}
