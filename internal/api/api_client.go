package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/schollz/progressbar/v3"
)

// AskOpenAi sends a prompt to the OpenAI API, processes the response stream and returns stats on it.
// Returns: timeToFirstToken (seconds), completionTokens, promptTokens, error
func AskOpenAi(client openai.Client, model string, prompt string, maxTokens int, bar *progressbar.ProgressBar) (string, float64, int, int, error) {
	start := time.Now()

	var (
		timeToFirstToken   float64
		firstTokenSeen     bool
		accumulatedContent string // Accumulate all content to count tokens more accurately
		estimatedTokens    int    // Real-time token estimation
		promptTokens       int
		completionTokens   int
	)

	// Create streaming request to the OpenAI API
	stream := client.Chat.Completions.NewStreaming(
		context.Background(),
		openai.ChatCompletionNewParams{
			Model: openai.ChatModel(model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(prompt),
			},
			MaxCompletionTokens: openai.Int(int64(maxTokens)),
			Temperature:         openai.Float(1.0),
			StreamOptions: openai.ChatCompletionStreamOptionsParam{
				IncludeUsage: openai.Bool(true),
			},
		},
	)

	// Process the stream
	for stream.Next() {
		event := stream.Current()

		// Check for usage information (it's only sent on the final message)
		if event.Usage.PromptTokens > 0 || event.Usage.CompletionTokens > 0 {
			promptTokens = int(event.Usage.PromptTokens)
			completionTokens = int(event.Usage.CompletionTokens)
		}

		// Process content from choices
		if len(event.Choices) > 0 {
			choice := event.Choices[0]

			// Track time to first token
			if !firstTokenSeen && choice.Delta.Content != "" {
				if strings.TrimSpace(choice.Delta.Content) != "" {
					timeToFirstToken = time.Since(start).Seconds()
					firstTokenSeen = true
				}
			}

			// Accumulate content and estimate tokens
			if choice.Delta.Content != "" {
				accumulatedContent += choice.Delta.Content

				// Estimate number of tokens in current chunk
				newTokens := estimateTokens(choice.Delta.Content)
				estimatedTokens += newTokens

				if bar != nil {
					bar.Add(newTokens)
				}
			}
		}
	}

	// Check for errors during streaming
	if err := stream.Err(); err != nil {
		return "", 0, 0, 0, fmt.Errorf("stream error: %w", err)
	}

	// If we got usage info from the stream, use it; otherwise use estimated tokens
	if completionTokens == 0 {
		completionTokens = estimatedTokens
	} else if bar != nil && completionTokens > 0 {
		// Final adjustment: if we have actual completion tokens, adjust the progress bar
		diff := completionTokens - estimatedTokens
		if diff != 0 { // Could be positive or negative
			bar.Add(diff)
		}
	}

	return accumulatedContent, timeToFirstToken, completionTokens, promptTokens, nil
}

func AskOpenAiRandomInput(client openai.Client, model string, numWords int, maxTokens int, bar *progressbar.ProgressBar) (string, float64, int, int, error) {
	prompt := generateRandomPhrase(numWords)
	return AskOpenAi(client, model, prompt, maxTokens, bar)
}

func estimateTokens(content string) int {
	if content == "" {
		return 0
	}

	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return 0
	}

	words := strings.Fields(content)
	wordCount := len(words)

	// Different strategies based on content type
	if wordCount > 0 {
		// For text with clear word boundaries: ~1.3 tokens per word on average
		// This accounts for subword tokenization in modern models
		return max(1, int(float64(wordCount)*1.3))
	} else {
		// For content without clear word boundaries (like punctuation, single chars)
		// Use character-based estimation: ~3-4 characters per token
		charCount := len(content)
		return max(1, int(float64(charCount)/3.0))
	}
}

// GetFirstAvailableModel retrieves the first available model from the OpenAI API.
func GetFirstAvailableModel(client openai.Client) (string, error) {
	// List models from the API
	modelList, err := client.Models.List(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to list models: %w", err)
	}

	// Check if there are any models available
	if len(modelList.Data) == 0 {
		return "", fmt.Errorf("no models available")
	}

	return modelList.Data[0].ID, nil
}
