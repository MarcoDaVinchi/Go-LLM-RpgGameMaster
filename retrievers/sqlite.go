package retrievers

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"go-llm-rpggamemaster/interfaces"
)

type SQLiteRetriever struct {
	db       *sql.DB
	embedder interfaces.VectorEmbeddingProvider
}

func NewSQLiteRetriever(embedder interfaces.VectorEmbeddingProvider) (*SQLiteRetriever, error) {
	return NewSQLiteRetrieverWithPath(embedder, "base.db")
}

func NewSQLiteRetrieverWithPath(embedder interfaces.VectorEmbeddingProvider, dbPath string) (*SQLiteRetriever, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Error().Err(err).Str("dbPath", dbPath).Msg("failed to open SQLite database")
		return nil, err
	}

	return &SQLiteRetriever{
		db:       db,
		embedder: embedder,
	}, nil
}

func (r *SQLiteRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]interfaces.Document, error) {
	return nil, nil
}

func (r *SQLiteRetriever) AddDocuments(ctx context.Context, docs []interfaces.Document) error {
	return nil
}
