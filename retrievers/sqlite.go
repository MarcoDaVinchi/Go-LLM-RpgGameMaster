package retrievers

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
)

type SQLiteRetriever struct {
	db *sql.DB
}

func NewSQLiteRetriever(embedder *embeddings.EmbedderImpl) (*SQLiteRetriever, error) {
	return NewSQLiteRetrieverWithPath(embedder, "base.db")
}

func NewSQLiteRetrieverWithPath(embedder *embeddings.EmbedderImpl, dbPath string) (*SQLiteRetriever, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Error().Err(err).Str("dbPath", dbPath).Msg("failed to open SQLite database")
		return nil, err
	}

	return &SQLiteRetriever{
		db: db,
	}, nil
}

func (r *SQLiteRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	// Stub implementation
	return nil, nil
}

func (r *SQLiteRetriever) AddDocuments(ctx context.Context, docs []schema.Document) error {
	// Stub implementation
	return nil
}
