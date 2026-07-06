package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"main/internal/db"
)

//go:embed sql/schema.sql
var schemaSQL string

func main() {
	dbPath := flag.String("db", "data/db/dnd.db", "path to SQLite database")
	spellsDir := flag.String("dir", "data/spells", "directory containing spells-*.json files")
	deleteDB := flag.Bool("delete", false, "delete the SQLite database file and exit")
	listSources := flag.Bool("list", false, "list all available spell source files in the spells directory and exit")
	filesToIngest := flag.String("files", "", "comma-separated list of spell files or source keys to ingest (e.g. 'phb,tce' or 'spells-phb.json')")
	flag.Parse()

	// Handle legacy positional argument for spells directory
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

	if *listSources {
		files, err := getAvailableSpellFiles(*spellsDir)
		if err != nil {
			log.Fatalf("failed to read spells directory: %v", err)
		}
		fmt.Println("Available spell sources:")
		for _, f := range files {
			base := filepath.Base(f)
			source := strings.TrimSuffix(strings.TrimPrefix(base, "spells-"), ".json")
			fmt.Printf("  - %-15s (file: %s)\n", source, base)
		}
		return
	}

	// Connect to the DB (this will also create the directory and execute schema.sql)
	conn, err := db.Connect(*dbPath, schemaSQL)
	if err != nil {
		log.Fatalf("failed to connect/init database: %v", err)
	}
	defer conn.Close()

	var targetFiles []string
	if *filesToIngest != "" {
		parts := strings.Split(*filesToIngest, ",")
		available, err := getAvailableSpellFiles(*spellsDir)
		if err != nil {
			log.Fatalf("failed to read spells directory: %v", err)
		}

		fileMap := make(map[string]string)
		for _, f := range available {
			base := filepath.Base(f)
			fileMap[strings.ToLower(base)] = f

			short := strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(base, "spells-"), ".json"))
			fileMap[short] = f
		}

		for _, part := range parts {
			part = strings.ToLower(strings.TrimSpace(part))
			if f, ok := fileMap[part]; ok {
				targetFiles = append(targetFiles, f)
			} else if f, ok := fileMap["spells-"+part+".json"]; ok {
				targetFiles = append(targetFiles, f)
			} else {
				log.Fatalf("unknown spell source or file: %s", part)
			}
		}
	} else {
		var err error
		targetFiles, err = getAvailableSpellFiles(*spellsDir)
		if err != nil {
			log.Fatalf("failed to scan spells: %v", err)
		}
	}

	if len(targetFiles) == 0 {
		log.Fatalf("no spells-*.json files found to ingest in %s", *spellsDir)
	}

	ctx := context.Background()
	total := 0
	for _, path := range targetFiles {
		n, err := db.IngestFile(ctx, conn, path)
		if err != nil {
			log.Fatalf("%s: %v", path, err)
		}
		fmt.Printf("%s: %d spells ingested\n", filepath.Base(path), n)
		total += n
	}
	fmt.Printf("done, %d spells total ingested\n", total)
}

func getAvailableSpellFiles(dir string) ([]string, error) {
	return filepath.Glob(filepath.Join(dir, "spells-*.json"))
}
