package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/bkovacki/gopenrouter"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable is required")
	}

	// Create a new client
	client := gopenrouter.New(apiKey)

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

	// Build chat completion request using the builder pattern
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

	// Print the response
	if len(response.Choices) > 0 {
		fmt.Printf("Assistant: %s\n", response.Choices[0].Message.Content)
		fmt.Printf("Finish Reason: %s\n", response.Choices[0].FinishReason)
	}

	// Print usage information if available
	if response.Usage.TotalTokens > 0 {
		fmt.Printf("Token Usage - Prompt: %d, Completion: %d, Total: %d\n",
			response.Usage.PromptTokens,
			response.Usage.CompletionTokens,
			response.Usage.TotalTokens)
	}

	// Example with provider preferences
	fmt.Println("\n--- Example with Provider Options ---")

	providerOptions := gopenrouter.NewProviderOptionsBuilder().
		WithAllowFallbacks(true).
		WithMaxPromptPrice(0.01).
		WithMaxCompletionPrice(0.02).
		Build()

	advancedRequest := gopenrouter.NewChatCompletionRequestBuilder("anthropic/claude-3-haiku", messages).
		WithProvider(providerOptions).
		WithMaxTokens(150).
		WithTemperature(0.5).
		WithTopP(0.9).
		WithFrequencyPenalty(0.1).
		WithPresencePenalty(0.1).
		Build()

	advancedResponse, err := client.ChatCompletion(ctx, *advancedRequest)
	if err != nil {
		log.Fatalf("Advanced chat completion failed: %v", err)
	}

	if len(advancedResponse.Choices) > 0 {
		fmt.Printf("Assistant (Advanced): %s\n", advancedResponse.Choices[0].Message.Content)
	}

	// Example multi-turn conversation
	fmt.Println("\n--- Multi-turn Conversation Example ---")

	conversationMessages := []gopenrouter.ChatMessage{
		{Role: "system", Content: "You are a helpful math tutor."},
		{Role: "user", Content: "What is 2 + 2?"},
		{Role: "assistant", Content: "2 + 2 equals 4."},
		{Role: "user", Content: "Now what is 4 * 3?"},
	}

	conversationRequest := gopenrouter.NewChatCompletionRequestBuilder("openai/gpt-4", conversationMessages).
		WithMaxTokens(50).
		WithTemperature(0.3).
		Build()

	conversationResponse, err := client.ChatCompletion(ctx, *conversationRequest)
	if err != nil {
		log.Fatalf("Conversation completion failed: %v", err)
	}

	if len(conversationResponse.Choices) > 0 {
		fmt.Printf("Math Tutor: %s\n", conversationResponse.Choices[0].Message.Content)
	}
}
