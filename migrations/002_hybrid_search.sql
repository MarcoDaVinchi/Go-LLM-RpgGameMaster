-- Migration: Hybrid Search Helper Functions
-- Description: SQL functions for semantic and keyword search
-- Dependencies: 001_initial_schema.sql

-- Create full-text search index on content
-- This enables efficient full-text search using tsquery
CREATE INDEX IF NOT EXISTS idx_content_fts ON context_items
USING GIN(to_tsvector('english', content));

-- Semantic search function
-- Performs vector similarity search using cosine distance
-- Returns ranked results for RRF fusion in Go
CREATE OR REPLACE FUNCTION semantic_search(
    query_embedding vector,
    p_game_id UUID,
    p_limit INTEGER DEFAULT 10
) RETURNS TABLE (
    id UUID,
    content TEXT,
    metadata JSONB,
    rank INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ci.id,
        ci.content,
        ci.metadata,
        ROW_NUMBER() OVER (ORDER BY ci.embedding <=> query_embedding)::INTEGER as rank
    FROM context_items ci
    WHERE ci.game_id = p_game_id
      AND ci.embedding IS NOT NULL
    ORDER BY ci.embedding <=> query_embedding
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Keyword search function
-- Performs full-text search using tsquery and ts_rank
-- Returns ranked results for RRF fusion in Go
CREATE OR REPLACE FUNCTION keyword_search(
    keyword_query TEXT,
    p_game_id UUID,
    p_limit INTEGER DEFAULT 10
) RETURNS TABLE (
    id UUID,
    content TEXT,
    metadata JSONB,
    rank INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ci.id,
        ci.content,
        ci.metadata,
        ROW_NUMBER() OVER (
            ORDER BY ts_rank(
                to_tsvector('english', ci.content),
                plainto_tsquery('english', keyword_query)
            ) DESC
        )::INTEGER as rank
    FROM context_items ci
    WHERE ci.game_id = p_game_id
      AND to_tsvector('english', ci.content) @@ plainto_tsquery('english', keyword_query)
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;