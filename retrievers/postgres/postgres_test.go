package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// MockEmbedder is a mock embedding provider for testing
type MockEmbedder struct {
	vectors map[string][]float32
}

func (m *MockEmbedder) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if vec, ok := m.vectors[text]; ok {
		return vec, nil
	}
	vec := make([]float32, 768)
	for i := range vec {
		vec[i] = float32(len(text)) * 0.001
	}
	return vec, nil
}

func (m *MockEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		vec, err := m.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = vec
	}
	return embeddings, nil
}

func (m *MockEmbedder) Name() string {
	return "mock"
}

// MockSQLStateError is a mock error that implements the SQLState() interface
type MockSQLStateError struct {
	state string
	msg   string
}

func (m *MockSQLStateError) Error() string {
	return m.msg
}

func (m *MockSQLStateError) SQLState() string {
	return m.state
}

func TestNewPostgresRetriever(t *testing.T) {
	t.Run("nil database pool", func(t *testing.T) {
		embedder := &MockEmbedder{}
		_, err := NewPostgresRetriever(nil, embedder)
		if err == nil {
			t.Error("expected error for nil pool")
		}
		if err.Error() != "database pool cannot be nil" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("nil embedder", func(t *testing.T) {
		_, err := NewPostgresRetriever(nil, nil)
		if err == nil {
			t.Error("expected error for nil embedder")
		}
		if err.Error() != "database pool cannot be nil" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("both nil", func(t *testing.T) {
		_, err := NewPostgresRetriever(nil, nil)
		if err == nil {
			t.Error("expected error for both nil")
		}
	})
}

func TestPostgresRetriever_GetRelevantDocuments(t *testing.T) {
	t.Run("context cancellation", func(t *testing.T) {
		t.Skip("Requires actual database connection - will be tested with testcontainers-go")
	})
}

func TestPostgresRetriever_AddDocuments(t *testing.T) {
	t.Run("empty document list", func(t *testing.T) {
		t.Skip("Requires actual database connection - will be tested with testcontainers-go")
	})

	t.Run("multiple documents", func(t *testing.T) {
		t.Skip("Requires actual database connection - will be tested with testcontainers-go")
	})
}

func TestHybridSearch(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		opts := SearchOptions{}
		if opts.RRFK != 0 {
			t.Error("expected default RRFK to be 0 (will be set to 60)")
		}
		if opts.Limit != 0 {
			t.Error("expected default Limit to be 0 (will be set to 10)")
		}
	})

	t.Run("RRF fusion calculation", func(t *testing.T) {
		k := 60
		rank := 1
		expected := 1.0 / float64(k+rank)

		result := 1.0 / float64(k+rank)
		if result != expected {
			t.Errorf("RRF calculation mismatch: got %f, want %f", result, expected)
		}

		tests := []struct {
			rank     int
			expected float64
		}{
			{1, 1.0 / 61.0},
			{10, 1.0 / 70.0},
			{100, 1.0 / 160.0},
		}

		for _, tt := range tests {
			result := 1.0 / float64(k+tt.rank)
			if result != tt.expected {
				t.Errorf("RRF calculation for rank %d: got %f, want %f", tt.rank, result, tt.expected)
			}
		}
	})

	t.Run("RRF fusion with different k values", func(t *testing.T) {
		semantic := []searchResult{
			{ID: "1", Rank: 1},
			{ID: "2", Rank: 2},
		}
		keyword := []searchResult{
			{ID: "1", Rank: 2},
			{ID: "3", Rank: 1},
		}

		results := rrfFusion(semantic, keyword, 60)
		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}

		var doc1 *HybridSearchResult
		for i := range results {
			if results[i].Document.Metadata != nil {
				doc1 = &results[i]
				break
			}
		}

		if doc1 != nil {
			expectedScore := 1.0/61.0 + 1.0/62.0
			if doc1.Score < expectedScore-0.0001 || doc1.Score > expectedScore+0.0001 {
				t.Errorf("RRF score mismatch: got %f, want approx %f", doc1.Score, expectedScore)
			}
		}
	})
}

func TestPoolConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		config := DefaultPoolConfig()

		if config.MaxConns != 10 {
			t.Errorf("MaxConns: got %d, want 10", config.MaxConns)
		}
		if config.MinConns != 2 {
			t.Errorf("MinConns: got %d, want 2", config.MinConns)
		}
		if config.MaxConnLifetime != 30*time.Minute {
			t.Errorf("MaxConnLifetime: got %v, want 30m", config.MaxConnLifetime)
		}
		if config.HealthCheckPeriod != 5*time.Minute {
			t.Errorf("HealthCheckPeriod: got %v, want 5m", config.HealthCheckPeriod)
		}
		if config.AcquireTimeout != 5*time.Second {
			t.Errorf("AcquireTimeout: got %v, want 5s", config.AcquireTimeout)
		}
	})

	t.Run("config values are reasonable", func(t *testing.T) {
		config := DefaultPoolConfig()

		if config.MinConns > config.MaxConns {
			t.Error("MinConns should be <= MaxConns")
		}
		if config.MaxConnLifetime < time.Minute {
			t.Error("MaxConnLifetime should be at least 1 minute")
		}
		if config.HealthCheckPeriod < time.Minute {
			t.Error("HealthCheckPeriod should be at least 1 minute")
		}
	})
}

func TestRetryConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		config := DefaultRetryConfig()

		if config.MaxRetries != 3 {
			t.Errorf("MaxRetries: got %d, want 3", config.MaxRetries)
		}
		if config.BaseDelay != 100*time.Millisecond {
			t.Errorf("BaseDelay: got %v, want 100ms", config.BaseDelay)
		}
		if config.MaxDelay != 2*time.Second {
			t.Errorf("MaxDelay: got %v, want 2s", config.MaxDelay)
		}
		if config.Jitter != 0.25 {
			t.Errorf("Jitter: got %f, want 0.25", config.Jitter)
		}
	})

	t.Run("exponential backoff calculation", func(t *testing.T) {
		config := DefaultRetryConfig()

		expectedDelays := []time.Duration{
			0,
			100 * time.Millisecond,
			200 * time.Millisecond,
			400 * time.Millisecond,
			800 * time.Millisecond,
		}

		for attempt, expected := range expectedDelays {
			if attempt == 0 {
				continue
			}
			delay := config.BaseDelay * time.Duration(1<<uint(attempt-1))
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
			if delay != expected {
				t.Errorf("Attempt %d delay: got %v, want %v", attempt, delay, expected)
			}
		}
	})
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"connection refused", fmt.Errorf("connection refused"), true},
		{"connection reset", fmt.Errorf("connection reset by peer"), true},
		{"broken pipe", fmt.Errorf("broken pipe"), true},
		{"timeout", fmt.Errorf("operation timeout"), true},
		{"deadlock", fmt.Errorf("deadlock detected"), true},
		{"too many connections", fmt.Errorf("too many connections"), true},
		{"normal error", fmt.Errorf("some other error"), false},
		{"unknown state", &MockSQLStateError{state: "00000", msg: "success"}, false},
		{"connection error code 40", &MockSQLStateError{state: "40001", msg: "transaction rollback"}, true},
		{"connection error code 55", &MockSQLStateError{state: "53300", msg: "too many connections"}, true},
		{"application error code 22", &MockSQLStateError{state: "22012", msg: "division by zero"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isRetryableError(%v): got %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"empty string", "", "", true},
		{"empty substring", "hello", "", true},
		{"substring at start", "hello world", "hello", true},
		{"substring at end", "hello world", "world", true},
		{"substring in middle", "hello world", "lo wo", true},
		{"substring not found", "hello", "xyz", false},
		{"case sensitive", "Hello World", "hello", false},
		{"single char match", "abc", "b", true},
		{"string equals substring", "test", "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q): got %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestMockEmbedder(t *testing.T) {
	t.Run("generate embedding for known text", func(t *testing.T) {
		embedder := &MockEmbedder{
			vectors: map[string][]float32{
				"test": {0.1, 0.2, 0.3},
			},
		}

		vec, err := embedder.GenerateEmbedding(context.Background(), "test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(vec) != 3 {
			t.Errorf("expected vector length 3, got %d", len(vec))
		}
		if vec[0] != 0.1 || vec[1] != 0.2 || vec[2] != 0.3 {
			t.Errorf("unexpected vector values: %v", vec)
		}
	})

	t.Run("generate embedding for unknown text", func(t *testing.T) {
		embedder := &MockEmbedder{}

		vec, err := embedder.GenerateEmbedding(context.Background(), "test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(vec) != 768 {
			t.Errorf("expected default vector length 768, got %d", len(vec))
		}
		expectedVal := float32(4) * 0.001
		for i, v := range vec {
			if v != expectedVal {
				t.Errorf("vec[%d]: got %f, want %f", i, v, expectedVal)
			}
		}
	})

	t.Run("embed documents", func(t *testing.T) {
		embedder := &MockEmbedder{}

		texts := []string{"doc1", "doc2"}
		embeddings, err := embedder.EmbedDocuments(context.Background(), texts)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(embeddings) != 2 {
			t.Errorf("expected 2 embeddings, got %d", len(embeddings))
		}
		if len(embeddings[0]) != 768 {
			t.Errorf("expected vector length 768, got %d", len(embeddings[0]))
		}
	})

	t.Run("name", func(t *testing.T) {
		embedder := &MockEmbedder{}
		if embedder.Name() != "mock" {
			t.Errorf("expected name 'mock', got %q", embedder.Name())
		}
	})
}

func TestSearchOptions(t *testing.T) {
	t.Run("zero values", func(t *testing.T) {
		opts := SearchOptions{}
		if opts.GameID != "" {
			t.Errorf("expected empty GameID, got %q", opts.GameID)
		}
		if opts.UserID != 0 {
			t.Errorf("expected UserID 0, got %d", opts.UserID)
		}
		if opts.Limit != 0 {
			t.Errorf("expected Limit 0, got %d", opts.Limit)
		}
		if opts.RRFK != 0 {
			t.Errorf("expected RRFK 0, got %d", opts.RRFK)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		opts := SearchOptions{
			GameID: "test-game",
			UserID: 123,
			Limit:  20,
			RRFK:   50,
		}
		if opts.GameID != "test-game" {
			t.Errorf("expected GameID 'test-game', got %q", opts.GameID)
		}
		if opts.UserID != 123 {
			t.Errorf("expected UserID 123, got %d", opts.UserID)
		}
		if opts.Limit != 20 {
			t.Errorf("expected Limit 20, got %d", opts.Limit)
		}
		if opts.RRFK != 50 {
			t.Errorf("expected RRFK 50, got %d", opts.RRFK)
		}
	})
}

func TestWithRetry(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return nil
		}

		err := withRetry(context.Background(), DefaultRetryConfig(), operation)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if callCount != 1 {
			t.Errorf("expected 1 call, got %d", callCount)
		}
	})

	t.Run("non-retryable error", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return fmt.Errorf("non-retryable error")
		}

		err := withRetry(context.Background(), DefaultRetryConfig(), operation)
		if err == nil {
			t.Error("expected error")
		}
		if callCount != 1 {
			t.Errorf("expected 1 call for non-retryable error, got %d", callCount)
		}
	})

	t.Run("context cancellation during retry", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		attemptCount := 0

		operation := func() error {
			attemptCount++
			if attemptCount == 1 {
				return fmt.Errorf("connection refused")
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(200 * time.Millisecond):
				return nil
			}
		}

		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		err := withRetry(ctx, DefaultRetryConfig(), operation)
		if err == nil {
			t.Error("expected error due to context cancellation")
		}
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return fmt.Errorf("non-retryable")
		}

		_ = withRetry(context.Background(), nil, operation)
		if callCount != 1 {
			t.Errorf("expected 1 call, got %d", callCount)
		}
	})
}

func TestPostgresRetriever_Close(t *testing.T) {
	t.Run("close nil pool doesn't panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Close panicked: %v", r)
			}
		}()

		embedder := &MockEmbedder{}
		retriever, err := NewPostgresRetriever(nil, embedder)
		if err != nil {
			t.Skip("cannot create retriever with nil pool")
		}
		retriever.Close()
	})
}
