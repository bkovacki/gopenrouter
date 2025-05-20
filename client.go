// Package gopenrouter provides a Go client for the OpenRouter API.
// OpenRouter is a unified API that provides access to various AI models.
package gopenrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	// openRouterAPIURL is the default base URL for the OpenRouter API.
	openRouterAPIURL = "https://openrouter.ai/api/v1"
)

// Client represents the OpenRouter client for making API requests.
// It holds API credentials and configuration for communicating with OpenRouter.
type Client struct {
	apiKey     string
	baseURL    string
	siteURL    string
	siteTitle  string
	httpClient HTTPDoer
}

// Option defines a client option function for modifying Client properties.
// These are used with the New constructor function to customize client behavior.
type Option func(*Client)

// HTTPDoer is an interface for making HTTP requests.
// It abstracts HTTP operations to allow users to provide custom HTTP clients with
// their own configuration (like custom timeouts, transport settings, or middleware).
// This interface matches http.Client's Do method, so *http.Client satisfies it directly.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// New creates a new OpenRouter client with the provided API key and optional customization options.
// By default, it uses the standard OpenRouter API URL and the default HTTP client.
func New(apiKey string, options ...Option) *Client {
	c := &Client{
		apiKey:     apiKey,
		baseURL:    openRouterAPIURL,
		httpClient: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	return c
}

// WithSiteURL sets the site URL that will be passed in HTTP-Referer header to the OpenRouter API.
// This is useful for attribution and tracking usage from different applications.
func WithSiteURL(siteURL string) Option {
	return func(c *Client) {
		c.siteURL = siteURL
	}
}

// WithSiteTitle sets the site title that will be passed in X-Title header to the OpenRouter API.
// This provides additional context about the origin of requests.
func WithSiteTitle(siteTitle string) Option {
	return func(c *Client) {
		c.siteTitle = siteTitle
	}
}

// WithHTTPClient sets a custom HTTP client for making requests.
// Users can provide their own http.Client (or any HTTPDoer implementation)
// to customize timeouts, transport settings, proxies, or add middleware for
// logging, metrics collection, or request/response manipulation.
func WithHTTPClient(httpClient HTTPDoer) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBaseURL sets a custom base URL for the OpenRouter API.
// This is primarily useful for testing or when using a proxy.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// requestOptions holds the configuration for an HTTP request.
// It encapsulates request body, headers, and URL parameters.
type requestOptions struct {
	body   any
	header http.Header
	params url.Values
}

// requestOption defines a function that modifies requestOptions.
// It follows the functional options pattern for configuring HTTP requests.
type requestOption func(*requestOptions)

// withBody sets the body for an HTTP request.
// The body can be any value that can be marshaled to JSON or an io.Reader.
func withBody(body any) requestOption {
	return func(args *requestOptions) {
		args.body = body
	}
}

// withContentType sets the Content-Type header for an HTTP request.
// This specifies the format of the request body.
func withContentType(contentType string) requestOption {
	return func(args *requestOptions) {
		args.header.Set("Content-Type", contentType)
	}
}

// withQueryParam adds a query parameter to the URL for an HTTP request.
func withQueryParam(name string, value string) requestOption {
	return func(args *requestOptions) {
		args.params.Set(name, value)
	}
}

// setCommonHeaders sets common headers for all OpenRouter API requests.
// These include authentication and attribution headers.
func (c *Client) setCommonHeaders(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	if c.siteURL != "" {
		req.Header.Set("HTTP-Referer", c.siteURL)
	}

	if c.siteTitle != "" {
		req.Header.Set("X-Title", c.siteTitle)
	}
}

// newRequest creates a new HTTP request with the given method, URL and options.
// It handles serialization of the request body and setting of common headers.
func (c *Client) newRequest(ctx context.Context, method, requestURL string, setters ...requestOption) (*http.Request, error) {
	// Default Options
	args := &requestOptions{
		body:   nil,
		header: make(http.Header),
		params: make(url.Values),
	}
	for _, setter := range setters {
		setter(args)
	}

	var bodyReader io.Reader

	if args.body != nil {
		if v, ok := args.body.(io.Reader); ok {
			bodyReader = v
		} else {
			var reqBytes []byte
			reqBytes, err := json.Marshal(args.body)
			if err != nil {
				return nil, err
			}
			bodyReader = bytes.NewBuffer(reqBytes)
		}
	}

	if len(args.params) > 0 {
		requestURL = fmt.Sprintf("%s?%s", requestURL, args.params.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, bodyReader)
	if err != nil {
		return nil, err
	}
	if args.header != nil {
		req.Header = args.header
	}

	c.setCommonHeaders(req)

	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// sendRequest sends an HTTP request and processes the response.
// It handles common error cases and deserializes the response body into the provided value.
func (c *Client) sendRequest(req *http.Request, v any) error {
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := res.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("error closing response body: %w", cerr)
		}
	}()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return c.handleErrorResp(res)
	}

	if v == nil {
		return nil
	}
	return json.NewDecoder(res.Body).Decode(v)
}

// handleErrorResp processes an error response from the API.
// It extracts error details from the response body and returns an appropriate error.
func (c *Client) handleErrorResp(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error, reading response body: %w", err)
	}
	var errRes ErrorResponse
	err = json.Unmarshal(body, &errRes)
	if err != nil || errRes.Error == nil {
		reqErr := &RequestError{
			HTTPStatus:     resp.Status,
			HTTPStatusCode: resp.StatusCode,
			Err:            err,
			Body:           body,
		}
		if errRes.Error != nil {
			reqErr.Err = errRes.Error
		}
		return reqErr
	}

	return errRes.Error
}

// fullURL builds a complete API URL by combining the base URL with the provided suffix.
// It ensures proper URL formatting by handling trailing slashes.
func (c *Client) fullURL(suffix string) string {
	baseURL := strings.TrimRight(c.baseURL, "/")
	return fmt.Sprintf("%s%s", baseURL, suffix)
}
