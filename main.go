package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"main/internal/db"
)

//go:embed sql/schema.sql
var schemaSQL string

func main() {
	dbPath := flag.String("db", "data/db/dnd.db", "path to SQLite database")
	spellsDir := flag.String("dir", "data/spells", "directory containing spells-*.json files")
	deleteDB := flag.Bool("delete", false, "delete the SQLite database file and exit")
	flag.Parse()

	// Handle positional argument for directory if provided (for compatibility)
	if flag.NArg() > 0 {
		*spellsDir = flag.Arg(0)
	}

	if *deleteDB {
		fmt.Printf("Deleting database at %s...\n", *dbPath)
		if err := db.DeleteDB(*dbPath); err != nil {
			log.Fatalf("failed to delete database: %v", err)
		}
		fmt.Println("Database deleted successfully.")
		return
	}

	// Connect to the DB (this will also create the directory and execute schema.sql)
	conn, err := db.Connect(*dbPath, schemaSQL)
	if err != nil {
		log.Fatalf("failed to connect/init database: %v", err)
	}
	defer conn.Close()

	files, err := filepath.Glob(filepath.Join(*spellsDir, "spells-*.json"))
	if err != nil {
		log.Fatalf("glob: %v", err)
	}
	if len(files) == 0 {
		log.Fatalf("no spells-*.json files found in %s", *spellsDir)
	}

	ctx := context.Background()
	total := 0
	for _, path := range files {
		n, err := db.IngestFile(ctx, conn, path)
		if err != nil {
			log.Fatalf("%s: %v", path, err)
		}
		fmt.Printf("%s: %d spells\n", filepath.Base(path), n)
		total += n
	}
	fmt.Printf("done, %d spells total\n", total)
}
