package output

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PrintBenchmarkHeader prints the benchmark header with details about the test.
func PrintBenchmarkHeader(modelName string, inputTokens int, maxTokens int, latency float64) {
	banner :=
		`
################################################################################################################
				          LLM API Throughput Benchmark
				    https://github.com/Yoosu-L/llmapibenchmark
					 Time：%s
################################################################################################################`

	fmt.Printf(banner+"\n", time.Now().UTC().Format("2006-01-02 15:04:05 UTC+0"))
	fmt.Printf("Input Tokens: %d\n", inputTokens)
	fmt.Printf("Output Tokens: %d\n", maxTokens)
	fmt.Printf("Test Model: %s\n", modelName)
	fmt.Printf("Latency: %.2f ms\n\n", latency)
}

// SaveResultsToMD saves the benchmark results to a Markdown file.
func SaveResultsToMD(results [][]interface{}, modelName string, inputTokens int, maxTokens int, latency float64) {
	// sanitize modelName to create a safe filename (replace path separators)
	safeModelName := strings.ReplaceAll(modelName, "/", "_")
	safeModelName = strings.ReplaceAll(safeModelName, "\\", "_")
	safeModelName = strings.TrimSpace(safeModelName)
	if safeModelName == "" {
		safeModelName = "model"
	}
	filename := fmt.Sprintf("API_Throughput_%s.md", safeModelName)
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		return
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("```\nInput Tokens: %d\n", inputTokens))
	file.WriteString(fmt.Sprintf("Output Tokens: %d\n", maxTokens))
	file.WriteString(fmt.Sprintf("Test Model: %s\n", modelName))
	file.WriteString(fmt.Sprintf("Latency: %.2f ms\n```\n\n", latency))
	file.WriteString("| Concurrency | Generation Throughput (tokens/s) |  Prompt Throughput (tokens/s) | Min TTFT (s) | Max TTFT (s) | Success Rate |\n")
	file.WriteString("|-------------|----------------------------------|-------------------------------|--------------|--------------|--------------|\n")

	for _, result := range results {
		concurrency := result[0].(int)
		generationSpeed := result[1].(float64)
		promptThroughput := result[2].(float64)
		minTTFT := result[3].(float64)
		maxTTFT := result[4].(float64)
		successRate := result[5].(float64)
		file.WriteString(fmt.Sprintf("| %11d | %32.2f | %29.2f | %12.2f | %12.2f | %11.2f%% |\n",
			concurrency,
			generationSpeed,
			promptThroughput,
			minTTFT,
			maxTTFT,
			successRate*100))
	}

	fmt.Printf("Results saved to: %s\n\n", filename)
}

// SaveModelOutputs saves the model generation outputs to a JSON file.
// Each concurrency level and its corresponding outputs are stored with metadata.
func SaveModelOutputs(outputs []string, modelName string, concurrency int, timestamp time.Time) (string, error) {
	// Create outputs directory if it doesn't exist
	outputDir := "model_outputs"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Sanitize model name for filename
	safeModelName := strings.ReplaceAll(modelName, "/", "_")
	safeModelName = strings.ReplaceAll(safeModelName, "\\", "_")
	safeModelName = strings.TrimSpace(safeModelName)
	if safeModelName == "" {
		safeModelName = "model"
	}

	// Create filename with timestamp and concurrency level
	timeStr := timestamp.Format("2006-01-02_15-04-05")
	filename := filepath.Join(outputDir, fmt.Sprintf("%s_concurrency_%d_%s.json", safeModelName, concurrency, timeStr))

	// Create output data structure
	data := map[string]interface{}{
		"model_name":  modelName,
		"concurrency": concurrency,
		"timestamp":   timestamp.UTC(),
		"outputs":     outputs,
		"count":       len(outputs),
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return "", fmt.Errorf("failed to write output file: %w", err)
	}

	return filename, nil
}
