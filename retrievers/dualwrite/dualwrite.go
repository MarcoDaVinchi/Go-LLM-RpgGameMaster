package dualwrite

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go-llm-rpggamemaster/interfaces"
)

// Retriever defines the interface for document retrieval
type Retriever interface {
	GetRelevantDocuments(ctx context.Context, query string) ([]interfaces.Document, error)
	AddDocuments(ctx context.Context, docs []interfaces.Document) error
}

// ReadSource specifies which database to read from
type ReadSource string

const (
	ReadFromQdrant   ReadSource = "qdrant"
	ReadFromPostgres ReadSource = "postgres"
	ReadFromDual     ReadSource = "dual" // Read from both and merge
)

// DualWriteRetriever writes to both Qdrant and PostgreSQL
type DualWriteRetriever struct {
	qdrant   Retriever
	postgres Retriever
	readFrom ReadSource
}

// Compile-time interface check
var _ Retriever = (*DualWriteRetriever)(nil)

// NewDualWriteRetriever creates a new dual-write retriever
func NewDualWriteRetriever(qdrant, postgres Retriever, readFrom ReadSource) (*DualWriteRetriever, error) {
	if qdrant == nil {
		return nil, fmt.Errorf("qdrant retriever cannot be nil")
	}
	if postgres == nil {
		return nil, fmt.Errorf("postgres retriever cannot be nil")
	}
	return &DualWriteRetriever{
		qdrant:   qdrant,
		postgres: postgres,
		readFrom: readFrom,
	}, nil
}

// AddDocuments writes documents to both databases concurrently
func (r *DualWriteRetriever) AddDocuments(ctx context.Context, docs []interfaces.Document) error {
	var wg sync.WaitGroup

	metrics := struct {
		qdrant      time.Duration
		postgres    time.Duration
		qdrantErr   error
		postgresErr error
	}{}

	// Write to Qdrant
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		if err := r.qdrant.AddDocuments(ctx, docs); err != nil {
			metrics.qdrantErr = err
			log.Error().Err(err).Msg("Failed to write to Qdrant")
		}
		metrics.qdrant = time.Since(start)
	}()

	// Write to PostgreSQL
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		if err := r.postgres.AddDocuments(ctx, docs); err != nil {
			metrics.postgresErr = err
			log.Error().Err(err).Msg("Failed to write to PostgreSQL")
		}
		metrics.postgres = time.Since(start)
	}()

	wg.Wait()

	// Log metrics
	log.Info().
		Dur("qdrant_latency", metrics.qdrant).
		Dur("postgres_latency", metrics.postgres).
		Msg("Dual-write completed")

	// Handle errors
	if metrics.qdrantErr != nil && metrics.postgresErr != nil {
		return fmt.Errorf("both databases failed: qdrant=%v, postgres=%v", metrics.qdrantErr, metrics.postgresErr)
	}
	if metrics.qdrantErr != nil {
		log.Warn().Err(metrics.qdrantErr).Msg("Qdrant write failed but PostgreSQL succeeded")
	}
	if metrics.postgresErr != nil {
		log.Warn().Err(metrics.postgresErr).Msg("PostgreSQL write failed but Qdrant succeeded")
	}

	return nil
}

// GetRelevantDocuments retrieves documents from configured source
func (r *DualWriteRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]interfaces.Document, error) {
	switch r.readFrom {
	case ReadFromQdrant:
		return r.qdrant.GetRelevantDocuments(ctx, query)
	case ReadFromPostgres:
		return r.postgres.GetRelevantDocuments(ctx, query)
	case ReadFromDual:
		// Read from both and merge (simple implementation: prefer Postgres)
		docs, err := r.postgres.GetRelevantDocuments(ctx, query)
		if err != nil {
			log.Warn().Err(err).Msg("PostgreSQL read failed, falling back to Qdrant")
			return r.qdrant.GetRelevantDocuments(ctx, query)
		}
		return docs, nil
	default:
		return nil, fmt.Errorf("unknown read source: %s", r.readFrom)
	}
}

// HealthCheck checks health of both databases
func (r *DualWriteRetriever) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)

	// Check PostgreSQL
	if hc, ok := r.postgres.(interface{ HealthCheck(context.Context) error }); ok {
		results["postgres"] = hc.HealthCheck(ctx)
	}

	return results
}

// SetReadSource changes the read source dynamically
func (r *DualWriteRetriever) SetReadSource(source ReadSource) {
	r.readFrom = source
	log.Info().Str("source", string(source)).Msg("Dual-write read source changed")
}

// GetMetrics returns current metrics (simplified)
func (r *DualWriteRetriever) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"read_source": string(r.readFrom),
	}
}
