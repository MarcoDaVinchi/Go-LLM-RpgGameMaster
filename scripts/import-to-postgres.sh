#!/bin/bash
set -e

DATABASE_URL="${DATABASE_URL:-postgres://rpg:secret@localhost:5432/gamedb}"
CSV_FILE="${CSV_FILE:-migrations/data.csv}"

if [ ! -f "$CSV_FILE" ]; then
    echo "Error: CSV file not found: $CSV_FILE"
    echo "Usage: $0 [csv-file]"
    exit 1
fi

echo "Importing data from $CSV_FILE to PostgreSQL..."
echo "Database: $DATABASE_URL"

CSV_ABS_PATH=$(cd "$(dirname "$CSV_FILE")" && pwd)/$(basename "$CSV_FILE")

psql "$DATABASE_URL" -c "
    BEGIN;

    COPY context_items (id, game_id, user_id, content, embedding, metadata, created_at)
    FROM '$CSV_ABS_PATH'
    WITH (FORMAT csv, HEADER true, NULL '');

    COMMIT;
"

echo "Import complete!"

COUNT=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM context_items;" | xargs)
echo "Total records in context_items: $COUNT"
