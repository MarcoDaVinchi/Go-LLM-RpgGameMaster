package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/rs/zerolog/log"

	"go-llm-rpggamemaster/interfaces"
)

const (
	defaultTableName = "context_items"
)

// Retriever defines the interface for document retrieval
type Retriever interface {
	GetRelevantDocuments(ctx context.Context, query string) ([]interfaces.Document, error)
	AddDocuments(ctx context.Context, docs []interfaces.Document) error
}

// PostgresRetriever implements Retriever using PostgreSQL with pgvector
type PostgresRetriever struct {
	db       *pgxpool.Pool
	embedder interfaces.VectorEmbeddingProvider
	table    string
}

// Compile-time interface check
var _ Retriever = (*PostgresRetriever)(nil)

// NewPostgresRetriever creates a new PostgreSQL retriever
func NewPostgresRetriever(db *pgxpool.Pool, embedder interfaces.VectorEmbeddingProvider) (*PostgresRetriever, error) {
	if db == nil {
		return nil, fmt.Errorf("database pool cannot be nil")
	}
	if embedder == nil {
		return nil, fmt.Errorf("embedder cannot be nil")
	}
	return &PostgresRetriever{
		db:       db,
		embedder: embedder,
		table:    defaultTableName,
	}, nil
}

// GetRelevantDocuments retrieves documents relevant to a query using hybrid search
func (r *PostgresRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]interfaces.Document, error) {
	start := time.Now()

	var docs []interfaces.Document

	err := withRetry(ctx, DefaultRetryConfig(), func() error {
		var err error
		docs, err = r.HybridSearch(ctx, query, SearchOptions{
			Limit: 10,
			RRFK:  60,
		})
		return err
	})
	if err != nil {
		log.Error().
			Dur("query_duration", time.Since(start)).
			Err(err).
			Msg("Retrieving relevant documents")
		return nil, fmt.Errorf("querying documents: %w", err)
	}

	log.Debug().
		Dur("query_duration", time.Since(start)).
		Int("result_count", len(docs)).
		Msg("Documents retrieved successfully")

	return docs, nil
}

// AddDocuments adds documents to the retriever
func (r *PostgresRetriever) AddDocuments(ctx context.Context, docs []interfaces.Document) error {
	start := time.Now()

	// Generate embeddings for all documents
	texts := make([]string, len(docs))
	for i, doc := range docs {
		texts[i] = doc.PageContent
	}

	embeddings, err := r.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return fmt.Errorf("generating embeddings: %w", err)
	}

	if len(embeddings) != len(docs) {
		return fmt.Errorf("embedding count mismatch: expected %d, got %d", len(docs), len(embeddings))
	}

	// Insert each document with its embedding
	insertSQL := fmt.Sprintf("INSERT INTO %s (game_id, user_id, content, embedding, metadata) VALUES ($1, $2, $3, $4, $5)", r.table)
	for i, doc := range docs {
		gameID, _ := doc.Metadata["game_id"].(string)
		userID, _ := doc.Metadata["user_id"].(string)

		err = withRetry(ctx, DefaultRetryConfig(), func() error {
			_, err := r.db.Exec(ctx, insertSQL, gameID, userID, doc.PageContent, pgvector.NewVector(embeddings[i]), doc.Metadata)
			return err
		})
		if err != nil {
			return fmt.Errorf("inserting document %d: %w", i, err)
		}
	}

	log.Info().
		Int("document_count", len(docs)).
		Dur("operation_duration", time.Since(start)).
		Msg("Documents added successfully")

	return nil
}

// Close closes the database connection pool
func (r *PostgresRetriever) Close() {
	log.Debug().Msg("closing postgres retriever connection pool")
	r.db.Close()
}
