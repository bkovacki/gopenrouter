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
			name: "Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{"data":[{"id":"meta-llama/llama-3.2-3b-instruct:free","name":"Meta Llama 3.2 3B Instruct (free)","created":1727276400,"description":"A lightweight, state-of-the-art conversational AI model","architecture":{"input_modalities":["text"],"output_modalities":["text"],"tokenizer":"Llama3","instruct_type":"llama3"},"top_provider":{"is_moderated":true,"context_length":131072,"max_completion_tokens":8192},"pricing":{"prompt":"0","completion":"0","image":"0","request":"0","input_cache_read":"0","input_cache_write":"0","web_search":"0","internal_reasoning":"0"},"context_length":131072,"hugging_face_id":"meta-llama/Llama-3.2-3B-Instruct","per_request_limits":{"requests_per_minute":20},"supported_parameters":["temperature","top_p","top_k","frequency_penalty","presence_penalty","repetition_penalty","min_p","top_a","seed","max_tokens","stop","response_format","tools","tool_choice"]}]}`)
			},
			expectErr:   false,
			expectCount: 1,
			expectFirst: "meta-llama/llama-3.2-3b-instruct:free",
		},
		{
			name: "APIError",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{"error": {"code": 400, "message": "Invalid API key"}}`)
			},
			expectErr:    true,
			expectAPIErr: true,
		},
		{
			name: "UnexpectedHTML",
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
