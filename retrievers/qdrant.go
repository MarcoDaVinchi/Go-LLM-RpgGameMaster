package retrievers

import (
	"context"
	"net/url"

	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/qdrant"
)

const (
	defaultCollection = "game_collection"
)

type Retriever interface {
	GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error)
	AddDocuments(ctx context.Context, docs []schema.Document) error
}

type QdrantRetriever struct {
	store qdrant.Store
}

func NewQdrantRetriever(embedder *embeddings.EmbedderImpl) (*QdrantRetriever, error) {
	qdrantUrl, err := url.Parse("http://localhost:6333")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse qdrant url")
		return nil, err
	}

	store, err := qdrant.New(
		qdrant.WithURL(*qdrantUrl),
		qdrant.WithCollectionName(defaultCollection),
		qdrant.WithEmbedder(embedder),
	)
	if err != nil {
		return nil, err
	}

	return &QdrantRetriever{
		store: store,
	}, nil
}

func (r *QdrantRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	retriever := vectorstores.ToRetriever(r.store, 10)
	return retriever.GetRelevantDocuments(ctx, query)
}

func (r *QdrantRetriever) AddDocuments(ctx context.Context, docs []schema.Document) error {
	_, err := r.store.AddDocuments(ctx, docs)
	return err
}
