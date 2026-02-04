# Data Migration Guide: Qdrant to PostgreSQL

This guide describes how to migrate data from Qdrant to PostgreSQL with pgvector extension.

## Prerequisites

- Qdrant running with data to migrate
- PostgreSQL 16+ with pgvector extension running
- Schema applied: `migrations/001_initial_schema.sql`
- Hybrid search functions: `migrations/002_hybrid_search.sql`

## Migration Steps

### Step 1: Export from Qdrant

Export all data from Qdrant using the scroll API:

```bash
./scripts/export-qdrant.sh
```

This will:
- Connect to Qdrant at `$QDRANT_URL` (default: http://localhost:6333)
- Export collection `game_collection`
- Save batches to `migrations/export/batch_*.json`

### Step 2: Convert Data Format

Convert Qdrant JSON export to PostgreSQL CSV format:

```bash
go run scripts/convert-qdrant-to-postgres.go migrations/export/ migrations/data.csv
```

This will:
- Read all batch files from the export directory
- Convert to CSV format with columns: id, game_id, user_id, content, embedding, metadata, created_at
- Output to `migrations/data.csv`

### Step 3: Import to PostgreSQL

Import the CSV data into PostgreSQL:

```bash
./scripts/import-to-postgres.sh migrations/data.csv
```

This will:
- Use PostgreSQL COPY command for efficient bulk import
- Import into `context_items` table
- Show final record count

### Step 4: Validate Migration

Verify the migration was successful:

```bash
./scripts/validate-migration.sh
```

This will:
- Compare record counts between Qdrant and PostgreSQL
- Check for duplicate IDs
- Show sample records

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `QDRANT_URL` | http://localhost:6333 | Qdrant API URL |
| `DATABASE_URL` | postgres://rpg:secret@localhost:5432/gamedb | PostgreSQL connection string |
| `COLLECTION` | game_collection | Qdrant collection name |

## Rollback Plan

If migration fails:

1. Data remains in Qdrant (source of truth)
2. Drop PostgreSQL table: `DROP TABLE context_items;`
3. Re-run schema: `psql $DATABASE_URL -f migrations/001_initial_schema.sql`
4. Re-attempt migration

## Zero-Downtime Migration

For production environments:

1. Enable dual-write mode in application config
2. Run migration to backfill historical data
3. Gradually shift read traffic from Qdrant to PostgreSQL
4. Once stable, disable Qdrant writes
5. Remove Qdrant dependency

See `retrievers/dualwrite/dualwrite.go` for dual-write implementation.

## Troubleshooting

### Export fails with "collection not found"
Ensure Qdrant is running and the collection exists:
```bash
curl http://localhost:6333/collections
```

### Import fails with "permission denied"
Ensure PostgreSQL can read the CSV file. The file must be readable by the postgres user.

### Vector dimension mismatch
Ensure the embedding dimension in the schema (768) matches your Qdrant vectors.
