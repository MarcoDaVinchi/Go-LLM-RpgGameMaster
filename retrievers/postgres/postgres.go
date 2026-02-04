package postgres

import (
	"context"
	"fmt"

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

// GetRelevantDocuments retrieves documents relevant to a query
func (r *PostgresRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]interfaces.Document, error) {
	// Generate embedding for the query
	embeddings, err := r.embedder.EmbedDocuments(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("generating embedding: %w", err)
	}

	if len(embeddings) == 0 || len(embeddings[0]) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	// Query using semantic search with pgvector
	rows, err := r.db.Query(ctx,
		"SELECT id, content, metadata FROM context_items WHERE embedding IS NOT NULL ORDER BY embedding <=> $1 LIMIT 10",
		pgvector.NewVector(embeddings[0]),
	)
	if err != nil {
		return nil, fmt.Errorf("querying documents: %w", err)
	}
	defer rows.Close()

	docs := make([]interfaces.Document, 0)
	for rows.Next() {
		var doc interfaces.Document
		var id string
		err := rows.Scan(&id, &doc.PageContent, &doc.Metadata)
		if err != nil {
			return nil, fmt.Errorf("scanning document: %w", err)
		}
		docs = append(docs, doc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("processing results: %w", err)
	}

	return docs, nil
}

// AddDocuments adds documents to the retriever
func (r *PostgresRetriever) AddDocuments(ctx context.Context, docs []interfaces.Document) error {
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

		_, err = r.db.Exec(ctx, insertSQL, gameID, userID, doc.PageContent, pgvector.NewVector(embeddings[i]), doc.Metadata)
		if err != nil {
			return fmt.Errorf("inserting document %d: %w", i, err)
		}
	}

	return nil
}

// Close closes the database connection pool
func (r *PostgresRetriever) Close() {
	log.Debug().Msg("closing postgres retriever connection pool")
	r.db.Close()
}
