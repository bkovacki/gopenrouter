# GOpenRouter

The GOpenRouter library provides convenient access to the [OpenRouter](https://openrouter.ai/) REST API from applications written in Go. OpenRouter is a unified API that provides access to various AI models from different providers, including OpenAI, Anthropic, Google, and more.

[![License](https://img.shields.io/github/license/bkovacki/gopenrouter)](https://github.com/bkovacki/gopenrouter/blob/main/LICENSE)
[![codecov](https://codecov.io/gh/bkovacki/gopenrouter/graph/badge.svg?token=vXQDEiWmJI)](https://codecov.io/gh/bkovacki/gopenrouter)

> **Note**: Support for streaming responses and chat completions will be added in future releases.

## Features

- Complete OpenRouter API coverage
- Builder pattern for constructing requests
- Customizable HTTP client with middleware support
- Proper error handling and detailed error types
- Context support for request cancellation and timeouts
- Comprehensive documentation

## Installation

```bash
go get github.com/bkovacki/gopenrouter
```

## Requirements

This library requires Go 1.24+.

## Usage

Import the package in your Go code:

```go
import "github.com/bkovacki/gopenrouter"
```

### Creating a client

To use this library, create a new client with your OpenRouter API key:

```go
client := gopenrouter.New("your-api-key")
```

You can also customize the client with optional settings:

```go
client := gopenrouter.New(
    "your-api-key",
    gopenrouter.WithSiteURL("https://yourapp.com"),
    gopenrouter.WithSiteTitle("Your App Name"),
    gopenrouter.WithHTTPClient(customHTTPClient),
)
```

### Generating Completions

```go
// Create a completion request using the builder pattern
request := gopenrouter.NewCompletionRequestBuilder(
    "anthropic/claude-3-opus-20240229",
    "Write a short poem about Go programming.",
).WithMaxTokens(150).
  WithTemperature(0.7).
  Build()

// Send the completion request
ctx := context.Background()
resp, err := client.Completion(ctx, request)
if err != nil {
    log.Fatalf("Completion error: %v", err)
}

// Use the response
fmt.Println(resp.Choices[0].Text)
```

### Advanced Provider Routing

OpenRouter allows you to customize how your requests are routed between different AI providers:

```go
// Create provider routing options
providerOptions := gopenrouter.NewProviderOptionsBuilder().
    WithDataCollection("deny").
    WithSort("price").
    WithOrder([]string{"Anthropic", "OpenAI"}).
    WithIgnore([]string{"Mistral"}).
    Build()

// Include provider options in your completion request
request := gopenrouter.NewCompletionRequestBuilder(
    "anthropic/claude-3-opus-20240229",
    "Write a short story about a robot learning to code.",
).WithProvider(providerOptions).
  Build()
```

### Checking Credits and Usage

```go
// Get your account credit information
credits, err := client.GetCredits(ctx)
if err != nil {
    log.Fatalf("Error getting credits: %v", err)
}

fmt.Printf("Total credits: %.2f\n", credits.TotalCredits)
fmt.Printf("Total usage: %.2f\n", credits.TotalUsage)
```

### Listing Available Models

```go
// Get a list of all available models
models, err := client.ListModels(ctx)
if err != nil {
    log.Fatalf("Error listing models: %v", err)
}

// Display model information
for _, model := range models {
    fmt.Printf("Model: %s\n", model.Name)
    fmt.Printf("  Description: %s\n", model.Description)
    fmt.Printf("  Context Length: %.0f tokens\n", model.ContextLength)
}
```

### Getting Generation Details

```go
// Get details about a specific generation by its ID
generationID := "gen_abc123"
generation, err := client.GetGeneration(ctx, generationID)
if err != nil {
    log.Fatalf("Error getting generation: %v", err)
}

fmt.Printf("Generation Cost: $%.6f\n", generation.TotalCost)
fmt.Printf("Prompt Tokens: %d\n", generation.TokensPrompt)
fmt.Printf("Completion Tokens: %d\n", generation.TokensCompletion)
```

## Error Handling

The library provides detailed error types for API errors and request errors:

```go
resp, err := client.Completion(ctx, request)
if err != nil {
    var apiErr *gopenrouter.APIError
    var reqErr *gopenrouter.RequestError
    
    if errors.As(err, &apiErr) {
        // Handle API-specific error
        fmt.Printf("API Error: %s (Code: %d)\n", apiErr.Message, apiErr.Code)
    } else if errors.As(err, &reqErr) {
        // Handle request error
        fmt.Printf("Request Error: %s (Status: %d)\n", reqErr.Error(), reqErr.HTTPStatusCode)
    } else {
        // Handle other errors
        fmt.Printf("Unexpected error: %v\n", err)
    }
    return
}
```

## Development

### Running Tests

```bash
make test
```

### Linting

```bash
make lint
```

### Coverage Report

```bash
make cover
make cover-html  # Opens coverage report in browser
```

## Contributing

Contributions to GOpenRouter are welcome! Please feel free to submit a Pull Request.

## Roadmap

- [ ] Chat completion API support
- [ ] Streaming support for completion requests
- [ ] Examples for common use cases
- [ ] Additional helper methods for advanced use cases

## License

This library is distributed under the MIT license. See the LICENSE file for more information.
