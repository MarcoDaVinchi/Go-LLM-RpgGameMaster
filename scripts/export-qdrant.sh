#!/bin/bash
set -e

QDRANT_URL="${QDRANT_URL:-http://localhost:6333}"
COLLECTION="${COLLECTION:-game_collection}"
OUTPUT_DIR="${OUTPUT_DIR:-./migrations/export}"

echo "Exporting Qdrant collection: $COLLECTION"
mkdir -p "$OUTPUT_DIR"

# Get all points from collection
echo "Fetching points..."

# Use scroll API to get all points
OFFSET=""
BATCH_SIZE=100
COUNT=0

while true; do
    if [ -z "$OFFSET" ]; then
        RESPONSE=$(curl -s -X POST "$QDRANT_URL/collections/$COLLECTION/points/scroll" \
            -H "Content-Type: application/json" \
            -d "{\"limit\": $BATCH_SIZE, \"with_payload\": true, \"with_vector\": true}")
    else
        RESPONSE=$(curl -s -X POST "$QDRANT_URL/collections/$COLLECTION/points/scroll" \
            -H "Content-Type: application/json" \
            -d "{\"limit\": $BATCH_SIZE, \"offset\": $OFFSET, \"with_payload\": true, \"with_vector\": true}")
    fi

    # Check for errors
    if echo "$RESPONSE" | grep -q '"status":"error"'; then
        echo "Error: $RESPONSE"
        exit 1
    fi

    # Save batch
    echo "$RESPONSE" > "$OUTPUT_DIR/batch_${COUNT}.json"

    # Extract next offset
    NEXT_OFFSET=$(echo "$RESPONSE" | grep -o '"next_page_offset":[0-9]*' | cut -d: -f2)

    POINTS_COUNT=$(echo "$RESPONSE" | grep -o '"points":\[' | wc -l)
    echo "Batch $COUNT: exported points"

    COUNT=$((COUNT + 1))

    if [ -z "$NEXT_OFFSET" ] || [ "$NEXT_OFFSET" = "null" ]; then
        break
    fi
    OFFSET="$NEXT_OFFSET"
done

echo "Export complete. $COUNT batches saved to: $OUTPUT_DIR"
