# Chat Completion Example

This example demonstrates how to use the gopenrouter library for conversational AI interactions using various chat completion models available through OpenRouter.

## Features Demonstrated

- **Basic Chat Completion**: Simple question-answer interactions
- **System Messages**: Setting up AI assistant behavior and context
- **Provider Options**: Advanced routing with cost controls and fallbacks
- **Multi-turn Conversations**: Maintaining conversation history and context
- **Usage Reporting**: Token usage statistics and monitoring
- **Model Flexibility**: Working with different AI models (OpenAI, Anthropic, etc.)
- **Parameter Tuning**: Temperature, top-p, penalties, and other generation controls

## Prerequisites

1. Go 1.24 or later
2. OpenRouter API key

## Setup

1. Set your OpenRouter API key as an environment variable:
   ```bash
   export OPENROUTER_API_KEY="your-api-key-here"
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Running the Example

```bash
go run chat_completion.go
```

## Example Output

```
Assistant: Paris is the capital of France.
Finish Reason: stop
Token Usage - Prompt: 25, Completion: 8, Total: 33

--- Example with Provider Options ---
Assistant (Advanced): Paris is the capital of France. It's a beautiful city known for its art, culture, cuisine, and iconic landmarks like the Eiffel Tower and the Louvre Museum.

--- Multi-turn Conversation Example ---
Math Tutor: 4 * 3 equals 12.
```

## Key Components

### 1. Basic Chat Completion
```go
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

request := gopenrouter.NewChatCompletionRequestBuilder("openai/gpt-3.5-turbo", messages).
    WithMaxTokens(100).
    WithTemperature(0.7).
    WithUsage(true).
    Build()
```

### 2. Provider Options and Cost Controls
```go
providerOptions := gopenrouter.NewProviderOptionsBuilder().
    WithAllowFallbacks(true).
    WithMaxPromptPrice(0.01).
    WithMaxCompletionPrice(0.02).
    Build()

request := gopenrouter.NewChatCompletionRequestBuilder("anthropic/claude-3-haiku", messages).
    WithProvider(providerOptions).
    WithMaxTokens(150).
    WithTemperature(0.5).
    WithTopP(0.9).
    WithFrequencyPenalty(0.1).
    WithPresencePenalty(0.1).
    Build()
```

### 3. Multi-turn Conversations
```go
conversationMessages := []gopenrouter.ChatMessage{
    {Role: "system", Content: "You are a helpful math tutor."},
    {Role: "user", Content: "What is 2 + 2?"},
    {Role: "assistant", Content: "2 + 2 equals 4."},
    {Role: "user", Content: "Now what is 4 * 3?"},
}
```

### 4. Making the Request
```go
ctx := context.Background()
response, err := client.ChatCompletion(ctx, request)

if len(response.Choices) > 0 {
    fmt.Printf("Assistant: %s\n", response.Choices[0].Message.Content)
    fmt.Printf("Finish Reason: %s\n", response.Choices[0].FinishReason)
}
```

## Message Roles

The chat completion API supports three types of message roles:

- **`system`**: Sets the behavior and context for the AI assistant
- **`user`**: Represents messages from the human user
- **`assistant`**: Represents responses from the AI assistant

## Supported Models

The example demonstrates usage with several popular models:

- **OpenAI Models**: `openai/gpt-3.5-turbo`, `openai/gpt-4`
- **Anthropic Models**: `anthropic/claude-3-haiku`
- **And many more**: OpenRouter provides access to dozens of AI models

## Configuration Parameters

### Generation Controls
- **`MaxTokens`**: Limits response length (e.g., 100 tokens)
- **`Temperature`**: Controls randomness (0.0 = deterministic, 2.0 = very random)
- **`TopP`**: Nucleus sampling parameter (0.0 to 1.0)
- **`TopK`**: Limits token selection to top K choices

### Repetition Controls
- **`FrequencyPenalty`**: Reduces repetition of token sequences (-2.0 to 2.0)
- **`PresencePenalty`**: Reduces repetition of topics (-2.0 to 2.0)
- **`RepetitionPenalty`**: General repetition penalty (0.0 to 2.0)

### Provider Options
- **`AllowFallbacks`**: Enable automatic fallback to alternative models
- **`MaxPromptPrice`**: Maximum cost per 1M prompt tokens
- **`MaxCompletionPrice`**: Maximum cost per 1M completion tokens
- **`RequireParameters`**: Ensure specific parameters are supported

## Usage Patterns

### Simple Q&A
Perfect for basic question-answering scenarios where context isn't needed.

### Conversational AI
Maintains conversation history to enable natural, contextual interactions.

### Specialized Assistants
Uses system messages to create domain-specific assistants (tutors, code reviewers, etc.).

### Cost-Conscious Applications
Implements price controls and fallbacks to manage API costs effectively.

## Error Handling

The example includes proper error handling patterns:

```go
response, err := client.ChatCompletion(ctx, request)
if err != nil {
    log.Fatalf("Chat completion failed: %v", err)
}
```

## Best Practices

1. **System Messages**: Always include a system message to set proper context
2. **Token Limits**: Set appropriate `MaxTokens` to control costs and response length
3. **Temperature Tuning**: Use lower temperatures (0.1-0.3) for factual responses, higher (0.7-1.0) for creative tasks
4. **Conversation History**: Include relevant previous messages for context
5. **Error Handling**: Always handle API errors gracefully
6. **Cost Monitoring**: Use provider options to control costs
7. **Model Selection**: Choose models appropriate for your use case and budget


## Next Steps

- Experiment with different models and parameters
- Add conversation persistence and memory management
- Integrate with web frameworks for chat applications
- Explore advanced features like function calling and structured outputs
- Try the simple completion example for non-conversational text generation
