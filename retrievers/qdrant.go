// Package qdrant provides a Qdrant-based vector retriever.
//
// Deprecated: Qdrant support is deprecated. Use PostgreSQL with pgvector instead.
// This package is kept for rollback safety but will be removed in a future release.
package retrievers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/rs/zerolog/log"
	"go-llm-rpggamemaster/interfaces"
)

const (
	defaultCollection = "game_collection"
)

type Retriever interface {
	GetRelevantDocuments(ctx context.Context, query string) ([]interfaces.Document, error)
	AddDocuments(ctx context.Context, docs []interfaces.Document) error
}

type QdrantRetriever struct {
	qdrantURL  string
	collection string
	embedder   interfaces.VectorEmbeddingProvider
	client     *http.Client
}

func NewQdrantRetriever(embedder interfaces.VectorEmbeddingProvider) (*QdrantRetriever, error) {
	qdrantURLEnv := os.Getenv("QDRANT_URL")
	if qdrantURLEnv == "" {
		log.Fatal().Msg("QDRANT_URL is not set")
	}

	_, err := url.Parse(qdrantURLEnv)
	if err != nil {
		log.Fatal().Err(err).Str("url", qdrantURLEnv).Msg("failed to parse qdrant url")
		return nil, err
	}

	return &QdrantRetriever{
		qdrantURL:  qdrantURLEnv,
		collection: defaultCollection,
		embedder:   embedder,
		client:     &http.Client{},
	}, nil
}

func (r *QdrantRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]interfaces.Document, error) {
	embeddings, err := r.embedder.EmbedDocuments(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	if len(embeddings) == 0 || len(embeddings[0]) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	reqBody := map[string]interface{}{
		"vector": embeddings[0],
		"limit":  10,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/points/search", r.qdrantURL, r.collection)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var searchResp struct {
		Result []struct {
			Payload map[string]interface{} `json:"payload"`
			Score   float32                `json:"score"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	docs := make([]interfaces.Document, 0, len(searchResp.Result))
	for _, result := range searchResp.Result {
		content, _ := result.Payload["content"].(string)
		docs = append(docs, interfaces.Document{
			PageContent: content,
			Metadata:    result.Payload,
		})
	}

	return docs, nil
}

func (r *QdrantRetriever) AddDocuments(ctx context.Context, docs []interfaces.Document) error {
	texts := make([]string, len(docs))
	for i, doc := range docs {
		texts[i] = doc.PageContent
	}

	embeddings, err := r.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return fmt.Errorf("embed documents: %w", err)
	}

	points := make([]map[string]interface{}, len(docs))
	for i, doc := range docs {
		points[i] = map[string]interface{}{
			"vector":  embeddings[i],
			"payload": doc.Metadata,
		}
	}

	reqBody := map[string]interface{}{
		"points": points,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/collections/%s/points", r.qdrantURL, r.collection)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
