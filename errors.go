package gopenrouter

import (
	"errors"
	"fmt"
)

var ErrCompletionStreamNotSupported = errors.New("streaming is not supported with this method")

// APIError provides error information returned by the OpenAI API.
type APIError struct {
	Code     int            `json:"code,omitempty"`
	Message  string         `json:"message"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// RequestError provides information about generic request errors.
type RequestError struct {
	HTTPStatus     string
	HTTPStatusCode int
	Err            error
	Body           []byte
}

type ErrorResponse struct {
	Error *APIError `json:"error,omitempty"`
}

func (e *APIError) Error() string {
	if e.Code > 0 {
		return fmt.Sprintf("error, status code: %d, message: %s, metadata: %v", e.Code, e.Message, e.Metadata)
	}
	return e.Message
}

func (e *RequestError) Error() string {
	return fmt.Sprintf(
		"error, status code: %d, status: %s, message: %s, body: %s",
		e.HTTPStatusCode, e.HTTPStatus, e.Err, e.Body,
	)
}

func (e *RequestError) Unwrap() error {
	return e.Err
}
