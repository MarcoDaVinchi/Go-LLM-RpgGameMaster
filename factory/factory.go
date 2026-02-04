package factory

import (
	"context"
	"fmt"
	"os"

	config "go-llm-rpggamemaster/config"
	factoryinterface "go-llm-rpggamemaster/factory/interface"
	"go-llm-rpggamemaster/interfaces"
	"go-llm-rpggamemaster/providers/routerai"
	"go-llm-rpggamemaster/retrievers"
	postgresretriever "go-llm-rpggamemaster/retrievers/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type providerFactory struct {
	cfg *config.Config
}

func NewProviderFactory(cfg *config.Config) factoryinterface.ProviderFactory {
	return &providerFactory{
		cfg,
	}
}

func (f *providerFactory) CreateInferenceProvider() (interfaces.InferenceProvider, error) {
	inferenceModel := f.cfg.InferenceModel
	baseURL := inferenceModel.Url
	apiKey := inferenceModel.ApiKey
	modelName := inferenceModel.Name
	providerType := inferenceModel.Type

	switch inferenceModel.Type {
	case config.ModelTypeRouterAI:
		return routerai.NewRouterAIProvider(modelName, apiKey, baseURL)
	case config.ModelTypeOpenAI, config.ModelTypeOllama:
		return nil, fmt.Errorf("provider type %s is deprecated, use routerai", providerType)
	default:
		return nil, fmt.Errorf("unsupported LLM provider type: %s", providerType)
	}
}

func (f *providerFactory) CreateEmbeddingProvider() (interfaces.VectorEmbeddingProvider, error) {
	embeddingModel := f.cfg.EmbeddingModel
	baseURL := embeddingModel.Url
	apiKey := embeddingModel.ApiKey
	modelName := embeddingModel.Name
	providerType := embeddingModel.Type

	switch providerType {
	case config.ModelTypeRouterAI:
		return routerai.NewRouterAIProvider(modelName, apiKey, baseURL)
	case config.ModelTypeOpenAI, config.ModelTypeOllama:
		return nil, fmt.Errorf("provider type %s is deprecated, use routerai", providerType)
	default:
		return nil, fmt.Errorf("unsupported embedding provider type: %s", providerType)
	}
}

func (f *providerFactory) CreateRetriever(embedder interfaces.VectorEmbeddingProvider, retrieverType string) (retrievers.Retriever, error) {
	switch retrieverType {
	case "sqlite":
		dbPath := "base.db"
		return retrievers.NewSQLiteRetrieverWithPath(embedder, dbPath)
	case "qdrant":
		return retrievers.NewQdrantRetriever(embedder)
	case "postgres":
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			return nil, fmt.Errorf("DATABASE_URL is not set")
		}
		pool, err := pgxpool.New(context.Background(), dbURL)
		if err != nil {
			return nil, fmt.Errorf("creating database pool: %w", err)
		}
		log.Info().Msg("PostgreSQL connection pool created")
		return postgresretriever.NewPostgresRetriever(pool, embedder)
	default:
		return nil, fmt.Errorf("unsupported retriever type: %s", retrieverType)
	}
}
