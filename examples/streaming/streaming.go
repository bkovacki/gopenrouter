package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/bkovacki/gopenrouter"
)

func main() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable is required")
	}

	client := gopenrouter.New(apiKey)
	ctx := context.Background()

	// Example 1: Streaming Completion
	fmt.Println("=== Streaming Completion Example ===")
	completionRequest := gopenrouter.NewCompletionRequestBuilder(
		"openai/gpt-3.5-turbo-instruct",
		"Write a short story about a robot discovering emotions:",
	).WithMaxTokens(200).Build()

	completionStream, err := client.CompletionStream(ctx, *completionRequest)
	if err != nil {
		log.Fatalf("Failed to create completion stream: %v", err)
	}
	defer completionStream.Close()

	fmt.Print("Response: ")
	for {
		chunk, err := completionStream.Recv()
		if err == io.EOF {
			fmt.Println("\n[Completion stream finished]")
			break
		}
		if err != nil {
			log.Fatalf("Error reading completion stream: %v", err)
		}

		// Process each choice in the chunk
		for _, choice := range chunk.Choices {
			if choice.Text != "" {
				fmt.Print(choice.Text)
			}
		}
	}

	fmt.Println()

	// Example 2: Streaming Chat Completion
	fmt.Println("=== Streaming Chat Completion Example ===")
	messages := []gopenrouter.ChatMessage{
		{Role: "system", Content: "You are a helpful assistant that writes creative stories."},
		{Role: "user", Content: "Tell me a brief story about a cat who learns to fly."},
	}

	chatRequest := gopenrouter.NewChatCompletionRequestBuilder(
		"openai/gpt-3.5-turbo",
		messages,
	).WithMaxTokens(150).WithTemperature(0.8).Build()

	chatStream, err := client.ChatCompletionStream(ctx, *chatRequest)
	if err != nil {
		log.Fatalf("Failed to create chat stream: %v", err)
	}
	defer chatStream.Close()

	fmt.Print("Assistant: ")
	for {
		chunk, err := chatStream.Recv()
		if err == io.EOF {
			fmt.Println("\n[Chat stream finished]")
			break
		}
		if err != nil {
			log.Fatalf("Error reading chat stream: %v", err)
		}

		// Process each choice in the chunk
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != nil {
				fmt.Printf("[C]%s[/C]", *choice.Delta.Content)
			}

			// Check if stream is finished
			if choice.FinishReason != nil {
				fmt.Printf("\n[Finished: %s]", *choice.FinishReason)
			}
		}
	}

	fmt.Println()

	// Example 3: Multiple models with streaming
	fmt.Println("=== Streaming with Model Fallback Example ===")
	multiModelRequest := gopenrouter.NewChatCompletionRequestBuilder(
		"openai/gpt-4", // Primary model
		[]gopenrouter.ChatMessage{
			{Role: "user", Content: "Explain quantum computing in simple terms."},
		},
	).WithModels([]string{
		"openai/gpt-4",
		"openai/gpt-3.5-turbo",     // Fallback
		"anthropic/claude-3-haiku", // Second fallback
	}).WithMaxTokens(100).Build()

	multiStream, err := client.ChatCompletionStream(ctx, *multiModelRequest)
	if err != nil {
		log.Fatalf("Failed to create multi-model stream: %v", err)
	}
	defer multiStream.Close()

	fmt.Print("Response: ")
	var modelUsed string
	for {
		chunk, err := multiStream.Recv()
		if err == io.EOF {
			fmt.Printf("\n[Stream finished using model: %s]", modelUsed)
			break
		}
		if err != nil {
			log.Fatalf("Error reading multi-model stream: %v", err)
		}

		// Track which model is being used
		if chunk.Model != "" {
			modelUsed = chunk.Model
		}

		// Process content
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != nil {
				fmt.Print(*choice.Delta.Content)
			}
		}
	}

	fmt.Println()
}
