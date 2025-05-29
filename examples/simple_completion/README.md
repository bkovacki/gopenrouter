# Simple Completion Example

This example demonstrates how to use the gopenrouter library to make simple text completion requests with comprehensive usage reporting and credits monitoring.

## Features Demonstrated

- **Simple Text Completion**: Basic completion request with prompt
- **Usage Reporting**: Enable usage statistics in the response
- **Generation Details**: Retrieve detailed metadata about the completion
- **Credits Monitoring**: Check account credits before and after requests
- **Provider Options**: Advanced routing with provider preferences
- **Cost Calculation**: Track the cost of individual requests

## Prerequisites

1. Go 1.19 or later
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
go run main.go
```

## Example Output

```
=== Initial Credits Status ===
Total Credits: $10.000000
Total Usage: $2.345678
Remaining: $7.654322

=== Simple Completion Example ===
Generated Text:  bright and full of possibilities. As we continue to develop more sophisticated AI systems, we can expect to see revolutionary changes in healthcare, education, transportation, and countless other fields. The key will be ensuring that these advances benefit all of humanity while addressing ethical concerns and potential risks.

Finish Reason: stop

Token Usage:
  Prompt Tokens: 8
  Completion Tokens: 52
  Total Tokens: 60

=== Generation Details (ID: gen_abc123def456) ===
Model Used: openai/gpt-3.5-turbo-instruct
Provider: openai
Total Cost: $0.000120
Usage: $0.000120
Cache Discount: $0.000000
Created At: 2024-01-15T10:30:45Z
Latency: 850 ms
Generation Time: 750 ms
Moderation Latency: 25 ms
Finish Reason: stop
Native Finish Reason: stop
Tokens - Prompt: 8, Completion: 52
Native Tokens - Prompt: 8, Completion: 52
Streamed: false
Cancelled: false
BYOK: false

=== Final Credits Status ===
Total Credits: $10.000000
Total Usage: $2.345798
Remaining: $7.654202
Cost of this request: $0.000120

=== Advanced Completion with Provider Options ===
Generated Haiku: Silicon minds think,
Algorithms dance with data—
Future unfolds bright.

Tokens Used: 25 (Prompt: 5, Completion: 20)

Example completed successfully!
```

## Key Components

### 1. Usage Reporting
```go
request := gopenrouter.NewCompletionRequestBuilder(model, prompt).
    WithUsage(true).
    Build()
```

### 2. Generation Details
```go
// Make the completion request
response, err := client.Completion(ctx, request)

// Get detailed generation information using the response ID
generationData, err := client.GetGeneration(ctx, response.ID)
```

### 3. Credits Monitoring
```go
// Check credits before the request
initialCredits, err := client.GetCredits(ctx)

// Make completion request...

// Check credits after the request
finalCredits, err := client.GetCredits(ctx)

// Calculate request cost
requestCost := finalCredits.TotalUsage - initialCredits.TotalUsage
```

### 4. Provider Options
```go
providerOptions := gopenrouter.NewProviderOptionsBuilder().
    WithAllowFallbacks(true).
    WithMaxPromptPrice(0.001).
    WithMaxCompletionPrice(0.002).
    Build()
```

## Understanding the Output

- **Credits Status**: Shows total credits purchased, total usage, and remaining balance
- **Token Usage**: Breakdown of prompt tokens, completion tokens, and total
- **Generation Details**: Comprehensive metadata including costs, latency, and provider information
- **Cost Tracking**: Calculates the cost of individual requests by comparing before/after credits
