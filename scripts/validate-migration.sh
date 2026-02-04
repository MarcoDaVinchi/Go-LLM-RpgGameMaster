#!/bin/bash
set -e

QDRANT_URL="${QDRANT_URL:-http://localhost:6333}"
DATABASE_URL="${DATABASE_URL:-postgres://rpg:secret@localhost:5432/gamedb}"
COLLECTION="${COLLECTION:-game_collection}"

echo "=== Migration Validation ==="
echo "Qdrant: $QDRANT_URL"
echo "PostgreSQL: $DATABASE_URL"
echo ""

# Get Qdrant count
echo "Fetching Qdrant count..."
QDRANT_RESPONSE=$(curl -s "$QDRANT_URL/collections/$COLLECTION")
QDRANT_COUNT=$(echo "$QDRANT_RESPONSE" | grep -o '"points_count":[0-9]*' | cut -d: -f2 || echo "0")

if [ -z "$QDRANT_COUNT" ] || [ "$QDRANT_COUNT" = "" ]; then
    QDRANT_COUNT=0
fi

# Get PostgreSQL count
echo "Fetching PostgreSQL count..."
PG_COUNT=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM context_items;" | xargs)

echo ""
echo "Results:"
echo "  Qdrant count: $QDRANT_COUNT"
echo "  PostgreSQL count: $PG_COUNT"
echo ""

# Compare counts
if [ "$QDRANT_COUNT" -eq "$PG_COUNT" ]; then
    echo "✓ VALIDATION PASSED: Counts match ($QDRANT_COUNT records)"
else
    echo "✗ VALIDATION FAILED: Count mismatch"
    echo "  Difference: $((QDRANT_COUNT - PG_COUNT))"
    exit 1
fi

# Check for duplicates
echo ""
echo "Checking for duplicates in PostgreSQL..."
DUPE_COUNT=$(psql "$DATABASE_URL" -t -c "
    SELECT COUNT(*) FROM (
        SELECT id FROM context_items
        GROUP BY id HAVING COUNT(*) > 1
    ) dupes;
" | xargs)

if [ "$DUPE_COUNT" -eq "0" ]; then
    echo "✓ No duplicates found"
else
    echo "✗ Found $DUPE_COUNT duplicate IDs"
    exit 1
fi

# Sample check
echo ""
echo "Sample records from PostgreSQL:"
psql "$DATABASE_URL" -c "
    SELECT id, LEFT(content, 50) as content_preview, created_at
    FROM context_items
    LIMIT 3;
"

echo ""
echo "✓ All validation checks passed!"
