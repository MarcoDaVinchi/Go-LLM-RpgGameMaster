package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// QdrantResponse represents the scroll API response
type QdrantResponse struct {
	Result struct {
		Points []QdrantPoint `json:"points"`
	} `json:"result"`
}

// QdrantPoint represents a single point from Qdrant
type QdrantPoint struct {
	ID      interface{}            `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <export-dir> [output.csv]\n", os.Args[0])
		os.Exit(1)
	}

	inputDir := os.Args[1]
	outputFile := "migrations/data.csv"
	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	}

	fmt.Printf("Converting Qdrant data from %s to %s\n", inputDir, outputFile)

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Create output file
	out, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	// Create CSV writer
	writer := csv.NewWriter(out)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"id", "game_id", "user_id", "content", "embedding", "metadata", "created_at"})

	// Process all batch files
	pattern := filepath.Join(inputDir, "batch_*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding batch files: %v\n", err)
		os.Exit(1)
	}

	var totalCount int
	for _, file := range files {
		count, err := processBatch(file, writer)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", file, err)
			continue
		}
		totalCount += count
		fmt.Printf("Processed %s: %d records\n", file, count)
	}

	fmt.Printf("Conversion complete. Total: %d records\n", totalCount)
}

func processBatch(filename string, writer *csv.Writer) (int, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	var response QdrantResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return 0, err
	}

	count := 0
	for _, point := range response.Result.Points {
		record := convertPointToRecord(point)
		if err := writer.Write(record); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing record: %v\n", err)
			continue
		}
		count++
	}

	return count, nil
}

func convertPointToRecord(point QdrantPoint) []string {
	// Extract ID
	id := ""
	switch v := point.ID.(type) {
	case string:
		id = v
	case float64:
		id = strconv.FormatInt(int64(v), 10)
	}

	// Extract payload fields
	gameID := getStringFromPayload(point.Payload, "game_id")
	userID := getStringFromPayload(point.Payload, "user_id")
	content := getStringFromPayload(point.Payload, "content")
	createdAt := getStringFromPayload(point.Payload, "created_at")

	if createdAt == "" {
		createdAt = "NOW()"
	}

	// Convert vector to PostgreSQL array format
	embedding := vectorToPostgresArray(point.Vector)

	// Convert remaining payload to JSON
	metadata, _ := json.Marshal(point.Payload)

	return []string{
		id,
		gameID,
		userID,
		content,
		embedding,
		string(metadata),
		createdAt,
	}
}

func getStringFromPayload(payload map[string]interface{}, key string) string {
	if val, ok := payload[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case float64:
			return strconv.FormatInt(int64(v), 10)
		}
	}
	return ""
}

func vectorToPostgresArray(vector []float32) string {
	if len(vector) == 0 {
		return "{}"
	}

	parts := make([]string, len(vector))
	for i, v := range vector {
		parts[i] = fmt.Sprintf("%f", v)
	}
	return "{" + strings.Join(parts, ",") + "}"
}
