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
	ctx := context.Background()

	// Check initial credits status
	fmt.Println("=== Initial Credits Status ===")
	creditsData, err := client.GetCredits(ctx)
	if err != nil {
		log.Printf("Warning: Could not retrieve credits: %v", err)
	} else {
		fmt.Printf("Total Credits: $%.6f\n", creditsData.TotalCredits)
		fmt.Printf("Total Usage: $%.6f\n", creditsData.TotalUsage)
		fmt.Printf("Remaining: $%.6f\n", creditsData.TotalCredits-creditsData.TotalUsage)
	}
	fmt.Println()

	// Simple completion example
	fmt.Println("=== Simple Completion Example ===")

	// Build completion request with usage reporting enabled
	request := gopenrouter.NewCompletionRequestBuilder(
		"openai/gpt-3.5-turbo-instruct",
		"The future of artificial intelligence is",
	).
		WithMaxTokens(100).
		WithTemperature(0.7).
		WithUsage(true).
		Build()

	// Make the completion request
	response, err := client.Completion(ctx, request)
	if err != nil {
		log.Fatalf("Completion failed: %v", err)
	}

	// Print the completion response
	if len(response.Choices) > 0 {
		fmt.Printf("Generated Text: %s\n", response.Choices[0].Text)
		fmt.Printf("Finish Reason: %s\n", response.Choices[0].FinishReason)
	}

	// Print usage information
	if response.Usage.TotalTokens > 0 {
		fmt.Printf("\nToken Usage:\n")
		fmt.Printf("  Prompt Tokens: %d\n", response.Usage.PromptTokens)
		fmt.Printf("  Completion Tokens: %d\n", response.Usage.CompletionTokens)
		fmt.Printf("  Total Tokens: %d\n", response.Usage.TotalTokens)
	}

	// Get detailed generation information using the response ID
	if response.ID != "" {
		fmt.Printf("\n=== Generation Details (ID: %s) ===\n", response.ID)
		generationData, err := client.GetGeneration(ctx, response.ID)
		if err != nil {
			log.Printf("Warning: Could not retrieve generation data: %v", err)
		} else {
			fmt.Printf("Model Used: %s\n", generationData.Model)
			fmt.Printf("Provider: %s\n", generationData.ProviderName)
			fmt.Printf("Total Cost: $%.6f\n", generationData.TotalCost)
			fmt.Printf("Usage: $%.6f\n", generationData.Usage)
			fmt.Printf("Cache Discount: $%.6f\n", generationData.CacheDiscount)
			fmt.Printf("Created At: %s\n", generationData.CreatedAt)
			fmt.Printf("Latency: %d ms\n", generationData.Latency)
			fmt.Printf("Generation Time: %d ms\n", generationData.GenerationTime)
			fmt.Printf("Moderation Latency: %d ms\n", generationData.ModerationLatency)
			fmt.Printf("Finish Reason: %s\n", generationData.FinishReason)
			fmt.Printf("Native Finish Reason: %s\n", generationData.NativeFinishReason)
			fmt.Printf("Tokens - Prompt: %d, Completion: %d\n",
				generationData.TokensPrompt, generationData.TokensCompletion)
			fmt.Printf("Native Tokens - Prompt: %d, Completion: %d\n",
				generationData.NativeTokensPrompt, generationData.NativeTokensCompletion)
			fmt.Printf("Streamed: %t\n", generationData.Streamed)
			fmt.Printf("Cancelled: %t\n", generationData.Cancelled)
			fmt.Printf("BYOK: %t\n", generationData.IsBYOK)
		}
	}

	// Check final credits status
	fmt.Println("\n=== Final Credits Status ===")
	finalCreditsData, err := client.GetCredits(ctx)
	if err != nil {
		log.Printf("Warning: Could not retrieve final credits: %v", err)
	} else {
		fmt.Printf("Total Credits: $%.6f\n", finalCreditsData.TotalCredits)
		fmt.Printf("Total Usage: $%.6f\n", finalCreditsData.TotalUsage)
		fmt.Printf("Remaining: $%.6f\n", finalCreditsData.TotalCredits-finalCreditsData.TotalUsage)

		// Calculate usage for this request
		if creditsData.TotalUsage > 0 {
			requestCost := finalCreditsData.TotalUsage - creditsData.TotalUsage
			if requestCost > 0 {
				fmt.Printf("Cost of this request: $%.6f\n", requestCost)
			}
		}
	}

	// Advanced completion example with provider options
	fmt.Println("\n=== Advanced Completion with Provider Options ===")

	providerOptions := gopenrouter.NewProviderOptionsBuilder().
		WithAllowFallbacks(true).
		WithMaxPromptPrice(0.01).
		WithMaxCompletionPrice(0.02).
		Build()

	advancedRequest := gopenrouter.NewCompletionRequestBuilder(
		"anthropic/claude-3-haiku",
		"Write a haiku about technology:",
	).
		WithProvider(providerOptions).
		WithMaxTokens(50).
		WithTemperature(0.8).
		WithTopP(0.9).
		WithFrequencyPenalty(0.1).
		WithPresencePenalty(0.1).
		WithUsage(true).
		Build()

	advancedResponse, err := client.Completion(ctx, advancedRequest)
	if err != nil {
		log.Printf("Advanced completion failed: %v", err)
	} else {
		if len(advancedResponse.Choices) > 0 {
			fmt.Printf("Generated Haiku: %s\n", advancedResponse.Choices[0].Text)
		}
		if advancedResponse.Usage.TotalTokens > 0 {
			fmt.Printf("Tokens Used: %d (Prompt: %d, Completion: %d)\n",
				advancedResponse.Usage.TotalTokens,
				advancedResponse.Usage.PromptTokens,
				advancedResponse.Usage.CompletionTokens)
		}
	}

	fmt.Println("\nExample completed successfully!")
}
