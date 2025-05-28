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

func TestClientCredits(t *testing.T) {
	cases := []struct {
		name         string
		handler      http.HandlerFunc
		expectErr    bool
		expectAPIErr bool
		expectReqErr bool
		expectTotal  float64
		expectUsage  float64
	}{
		{
			name: "Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{"data": {"total_credits": 42.5, "total_usage": 10.25}}`)
			},
			expectErr:   false,
			expectTotal: 42.5,
			expectUsage: 10.25,
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
			data, err := client.GetCredits(context.Background())

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
				if data.TotalCredits != tc.expectTotal {
					t.Errorf("unexpected total credits: got %v, want %v", data.TotalCredits, tc.expectTotal)
				}
				if data.TotalUsage != tc.expectUsage {
					t.Errorf("unexpected total usage: got %v, want %v", data.TotalUsage, tc.expectUsage)
				}
			}
		})
	}
}
