# Streaming Example

This example demonstrates how to use the streaming functionality of the gopenrouter client for real-time response generation.

## Features Demonstrated

- **Completion Streaming**: Stream text completions in real-time
- **Chat Completion Streaming**: Stream chat responses with role-based conversations
- **Model Fallback**: Use multiple models with automatic fallback for streaming
- **Error Handling**: Proper stream lifecycle management and error handling

## Prerequisites

Set your OpenRouter API key as an environment variable:

```bash
export OPENROUTER_API_KEY="your-api-key-here"
```

## Running the Example

```bash
go run main.go
```

## Key Concepts

### Streaming vs Non-Streaming

- **Non-streaming methods**: `client.Completion()` and `client.ChatCompletion()` return complete responses
- **Streaming methods**: `client.CompletionStream()` and `client.ChatCompletionStream()` return stream readers for real-time chunks

### Stream Reader Interface

Both streaming methods return readers that implement:

```go
type StreamReader[T any] interface {
    Recv() (T, error)  // Read next chunk
    Close() error      // Close stream and cleanup
}
```

### Processing Streaming Responses

#### Completion Streaming

```go
stream, err := client.CompletionStream(ctx, request)
defer stream.Close()

for {
    chunk, err := stream.Recv()
    if err == io.EOF {
        break // Stream finished
    }
    if err != nil {
        // Handle error
    }
    
    // Process chunk.Choices[].Delta content
}
```

#### Chat Completion Streaming

```go
stream, err := client.ChatCompletionStream(ctx, request)
defer stream.Close()

for {
    chunk, err := stream.Recv()
    if err == io.EOF {
        break // Stream finished
    }
    if err != nil {
        // Handle error
    }
    
    // Process chunk.Choices[].Delta.Content
}
```

### Best Practices

1. **Always close streams**: Use `defer stream.Close()` to ensure proper cleanup
2. **Handle EOF properly**: `io.EOF` indicates successful stream completion
3. **Check for finish reasons**: Monitor `choice.FinishReason` to understand why generation stopped
4. **Buffer content**: Accumulate delta content to build complete messages
5. **Error handling**: Handle network errors and malformed responses gracefully

### Stream Cancellation

Streams respect context cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

stream, err := client.ChatCompletionStream(ctx, request)
// Stream will be cancelled if context times out
```

### Model Selection

When using multiple models, the stream will indicate which model is actually being used:

```go
for {
    chunk, err := stream.Recv()
    // chunk.Model contains the actual model being used
}
```

This is particularly useful when using model fallbacks, as you can track which model in your list is responding to the request.