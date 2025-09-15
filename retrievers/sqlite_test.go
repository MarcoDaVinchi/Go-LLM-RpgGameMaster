package retrievers

import (
	"log"
	"os"
	"testing"
)

func TestSQLiteRetrieverConnectionWithoutTables(t *testing.T) {
	// Clean up any existing base.db
	os.Remove("base.db")
	defer os.Remove("base.db")

	// Create retriever directly (embedder not used in current implementation)
	retriever, err := NewSQLiteRetrieverWithPath(nil, "base.db")
	if err != nil {
		t.Fatalf("Failed to create SQLite retriever: %v", err)
	}

	// Check connection
	if retriever.db == nil {
		t.Fatal("Database connection is nil")
	}

	err = retriever.db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	log.Printf("Connection established successfully to base.db")

	// Check no tables exist
	rows, err := retriever.db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			t.Fatalf("Failed to scan table name: %v", err)
		}
		tables = append(tables, name)
	}

	if len(tables) != 0 {
		t.Fatalf("Expected no tables, but found: %v", tables)
	}

	log.Printf("Confirmed no tables exist in the database")
}
