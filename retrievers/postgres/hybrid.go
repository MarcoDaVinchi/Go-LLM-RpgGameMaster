package postgres

import (
	"context"
	"fmt"
	"sort"

	"github.com/pgvector/pgvector-go"

	"go-llm-rpggamemaster/interfaces"
)

// HybridSearchResult represents a single search result with RRF score
type HybridSearchResult struct {
	Document     interfaces.Document
	Score        float64
	SemanticRank int
	KeywordRank  int
}

// SearchOptions contains options for hybrid search
type SearchOptions struct {
	GameID string
	UserID int64
	Limit  int
	RRFK   int // RRF constant, default 60
}

// HybridSearch performs hybrid search combining semantic and keyword results
func (r *PostgresRetriever) HybridSearch(ctx context.Context, query string, opts SearchOptions) ([]interfaces.Document, error) {
	if opts.RRFK == 0 {
		opts.RRFK = 60
	}
	if opts.Limit == 0 {
		opts.Limit = 10
	}

	// Generate embedding for semantic search
	embeddings, err := r.embedder.EmbedDocuments(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("generating embedding: %w", err)
	}

	if len(embeddings) == 0 || len(embeddings[0]) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	embedding := embeddings[0]

	// Run semantic and keyword searches concurrently
	type searchResultWithErr struct {
		results []searchResult
		err     error
	}

	semanticChan := make(chan searchResultWithErr, 1)
	keywordChan := make(chan searchResultWithErr, 1)

	go func() {
		results, err := r.semanticSearch(ctx, embedding, opts.GameID, opts.UserID, opts.Limit*2)
		semanticChan <- searchResultWithErr{results: results, err: err}
	}()

	go func() {
		results, err := r.keywordSearch(ctx, query, opts.GameID, opts.UserID, opts.Limit*2)
		keywordChan <- searchResultWithErr{results: results, err: err}
	}()

	// Collect results
	var semantic []searchResult
	var keyword []searchResult

	for i := 0; i < 2; i++ {
		select {
		case res := <-semanticChan:
			if res.err != nil {
				return nil, res.err
			}
			semantic = res.results
		case res := <-keywordChan:
			if res.err != nil {
				return nil, res.err
			}
			keyword = res.results
		}
	}

	// Combine using RRF
	fused := rrfFusion(semantic, keyword, opts.RRFK)

	// Sort by score and limit
	sort.Slice(fused, func(i, j int) bool {
		return fused[i].Score > fused[j].Score
	})

	if len(fused) > opts.Limit {
		fused = fused[:opts.Limit]
	}

	// Convert to documents
	docs := make([]interfaces.Document, len(fused))
	for i, res := range fused {
		docs[i] = res.Document
	}

	return docs, nil
}

type searchResult struct {
	ID       string
	Content  string
	Metadata map[string]interface{}
	Rank     int
}

func (r *PostgresRetriever) semanticSearch(ctx context.Context, embedding []float32, gameID string, userID int64, limit int) ([]searchResult, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, content, metadata
		FROM context_items
		WHERE game_id = $1 AND user_id = $2 AND embedding IS NOT NULL
		ORDER BY embedding <=> $3
		LIMIT $4
	`, gameID, userID, pgvector.NewVector(embedding), limit)
	if err != nil {
		return nil, fmt.Errorf("semantic search query: %w", err)
	}
	defer rows.Close()

	var results []searchResult
	rank := 1
	for rows.Next() {
		var sr searchResult
		err := rows.Scan(&sr.ID, &sr.Content, &sr.Metadata)
		if err != nil {
			return nil, fmt.Errorf("scanning semantic result: %w", err)
		}
		sr.Rank = rank
		rank++
		results = append(results, sr)
	}
	return results, rows.Err()
}

func (r *PostgresRetriever) keywordSearch(ctx context.Context, query, gameID string, userID int64, limit int) ([]searchResult, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, content, metadata,
		       ts_rank(to_tsvector('english', content), plainto_tsquery('english', $1)) as rank
		FROM context_items
		WHERE game_id = $2 AND user_id = $3
		  AND to_tsvector('english', content) @@ plainto_tsquery('english', $1)
		ORDER BY rank DESC
		LIMIT $4
	`, query, gameID, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("keyword search query: %w", err)
	}
	defer rows.Close()

	var results []searchResult
	rank := 1
	for rows.Next() {
		var sr searchResult
		var tsRank float64
		err := rows.Scan(&sr.ID, &sr.Content, &sr.Metadata, &tsRank)
		if err != nil {
			return nil, fmt.Errorf("scanning keyword result: %w", err)
		}
		sr.Rank = rank
		rank++
		results = append(results, sr)
	}
	return results, rows.Err()
}

func rrfFusion(semantic, keyword []searchResult, k int) []HybridSearchResult {
	semanticMap := make(map[string]int)
	for _, s := range semantic {
		semanticMap[s.ID] = s.Rank
	}

	keywordMap := make(map[string]int)
	for _, kw := range keyword {
		keywordMap[kw.ID] = kw.Rank
	}

	// Collect all unique IDs
	allIDs := make(map[string]struct{})
	for _, s := range semantic {
		allIDs[s.ID] = struct{}{}
	}
	for _, kw := range keyword {
		allIDs[kw.ID] = struct{}{}
	}

	// Calculate RRF scores
	var results []HybridSearchResult
	for id := range allIDs {
		semanticRank := semanticMap[id]
		keywordRank := keywordMap[id]

		score := 0.0
		if semanticRank > 0 {
			score += 1.0 / float64(k+semanticRank)
		}
		if keywordRank > 0 {
			score += 1.0 / float64(k+keywordRank)
		}

		// Get document from either list
		var doc interfaces.Document
		for _, s := range semantic {
			if s.ID == id {
				doc = interfaces.Document{
					PageContent: s.Content,
					Metadata:    s.Metadata,
				}
				break
			}
		}
		if doc.PageContent == "" {
			for _, kw := range keyword {
				if kw.ID == id {
					doc = interfaces.Document{
						PageContent: kw.Content,
						Metadata:    kw.Metadata,
					}
					break
				}
			}
		}

		results = append(results, HybridSearchResult{
			Document:     doc,
			Score:        score,
			SemanticRank: semanticRank,
			KeywordRank:  keywordRank,
		})
	}

	return results
}
