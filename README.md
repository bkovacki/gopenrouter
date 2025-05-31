# GOpenRouter

The GOpenRouter library provides convenient access to the [OpenRouter](https://openrouter.ai/) REST API from applications written in Go. OpenRouter is a unified API that provides access to various AI models from different providers, including OpenAI, Anthropic, Google, and more.

[![License](https://img.shields.io/github/license/bkovacki/gopenrouter)](https://github.com/bkovacki/gopenrouter/blob/main/LICENSE)
[![codecov](https://codecov.io/gh/bkovacki/gopenrouter/graph/badge.svg?token=vXQDEiWmJI)](https://codecov.io/gh/bkovacki/gopenrouter)

> 🚀 **New Feature**: **Real-time streaming support** is now available for both completion and chat endpoints! Build interactive AI applications with live response generation. See the [**Streaming Documentation**](STREAMING.md) for complete details and examples.

## Quick Navigation

| What you want to do | Go to |
|---------------------|-------|
| 🚀 Get started quickly | [Installation](#installation) → [Usage](#usage) |
| 💬 Build chat applications | [Chat Completions](#chat-completions) |
| ⚡ Implement real-time streaming | [Streaming Documentation](STREAMING.md) |
| 🔧 Advanced configuration | [Advanced Provider Routing](#advanced-provider-routing) |
| 📊 Monitor usage and costs | [Checking Credits](#checking-credits-and-usage) |
| 🎯 See working examples | [Examples](#examples) |
| 🐛 Handle errors properly | [Error Handling](#error-handling) |

## Features

- Complete OpenRouter API coverage
- Text completion and chat completion support
- **Real-time streaming support** for both completion and chat endpoints
- Builder pattern for constructing requests
- Customizable HTTP client with middleware support
- Proper error handling and detailed error types
- Context support for request cancellation and timeouts
- Comprehensive documentation and examples

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

### Text Completions

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

### Chat Completions

```go
// Create conversation messages
messages := []gopenrouter.ChatMessage{
    {
        Role:    "system",
        Content: "You are a helpful assistant that provides concise answers.",
    },
    {
        Role:    "user",
        Content: "What is the capital of France?",
    },
}

// Build chat completion request
request := gopenrouter.NewChatCompletionRequestBuilder("openai/gpt-3.5-turbo", messages).
    WithMaxTokens(100).
    WithTemperature(0.7).
    WithUsage(true).
    Build()

// Make the chat completion request
ctx := context.Background()
response, err := client.ChatCompletion(ctx, *request)
if err != nil {
    log.Fatalf("Chat completion failed: %v", err)
}

// Use the response
fmt.Printf("Assistant: %s\n", response.Choices[0].Message.Content)
```

### Streaming Responses

The library provides comprehensive real-time streaming support for both completion and chat completion endpoints. Streaming allows you to:

- **Reduce perceived latency** by displaying responses as they are generated
- **Build interactive chat interfaces** with real-time feedback
- **Handle long responses efficiently** without waiting for complete generation
- **Implement live AI-powered features** with immediate user feedback

Quick streaming example:

```go
import (
    "context"
    "fmt"
    "io"
    "log"
    
    "github.com/bkovacki/gopenrouter"
)

// Streaming chat completion
messages := []gopenrouter.ChatMessage{
    {Role: "user", Content: "Tell me a story"},
}

request := gopenrouter.NewChatCompletionRequestBuilder("openai/gpt-3.5-turbo", messages).Build()

stream, err := client.ChatCompletionStream(ctx, *request)
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

fmt.Print("Assistant: ")
for {
    chunk, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    
    for _, choice := range chunk.Choices {
        if choice.Delta.Content != nil {
            fmt.Print(*choice.Delta.Content)
        }
    }
}
fmt.Println()
```

**📚 For comprehensive streaming documentation including:**
- Complete API reference and response types
- Advanced usage patterns and best practices  
- Error handling and resource management
- Performance considerations and limitations
- Migration guide from non-streaming code

See the [**Streaming Documentation**](STREAMING.md) for complete details and examples.

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

## Examples

The library includes comprehensive examples to help you get started:

> **⚠️ Cost Warning**: Running these examples will make actual API calls to OpenRouter and will incur charges based on your usage. Please monitor your credits and usage to avoid unexpected costs.

### Simple Completion Example
Located in `examples/simple_completion/`, this example demonstrates:
- Basic text completion with usage reporting
- Generation details retrieval using the generation endpoint
- Credits status monitoring before and after requests
- Cost calculation for individual requests
- Advanced provider options with cost controls

Run the example:
```bash
export OPENROUTER_API_KEY="your-api-key-here"
go run examples/simple_completion/simple_completion.go  # Note: This will incur API charges
```

### Chat Completion Example
Located in `examples/chat_completion/`, this example demonstrates:
- Basic chat completion with system and user messages
- Multi-turn conversations with context
- Provider options for cost control and fallbacks
- Different AI models (OpenAI, Anthropic, etc.)
- Parameter tuning (temperature, penalties, etc.)

Run the example:
```bash
export OPENROUTER_API_KEY="your-api-key-here"
go run examples/chat_completion/chat_completion.go  # Note: This will incur API charges
```

### Streaming Example
Located in `examples/streaming/`, this example demonstrates:
- Real-time streaming for both completion and chat completion
- Handling streaming responses and delta content
- Model fallback with streaming support
- Proper stream lifecycle management and error handling

Run the example:
```bash
export OPENROUTER_API_KEY="your-api-key-here"
go run examples/streaming/streaming.go  # Note: This will incur API charges
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

- [x] Chat completion API support
- [x] Examples for common use cases
- [x] Streaming support for completion and chat completion requests
- [ ] Additional helper methods for advanced use cases

## License

This library is distributed under the MIT license. See the LICENSE file for more information.
