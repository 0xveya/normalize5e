package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
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
	listSources := flag.Bool("list", false, "list all available ingestion targets and exit")
	filesToIngest := flag.String("files", "", "comma-separated list of targets to ingest (e.g. 'languages', 'conditions', 'phb', 'tce')")
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
		fmt.Println("Available general ingestion targets:")
		fmt.Println("  - languages       (file: data/languages.json)")
		fmt.Println("  - conditions      (file: data/conditionsdiseases.json)")
		fmt.Println("  - spells          (all spell JSON files in spells directory)")
		fmt.Println()

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

	shouldIngestLanguages := false
	shouldIngestConditions := false
	shouldIngestSpells := false
	var targetSpellFiles []string

	if *filesToIngest != "" {
		parts := strings.Split(*filesToIngest, ",")
		availableSpells, err := getAvailableSpellFiles(*spellsDir)
		if err != nil {
			log.Fatalf("failed to scan spells: %v", err)
		}

		spellFileMap := make(map[string]string)
		for _, f := range availableSpells {
			base := filepath.Base(f)
			spellFileMap[strings.ToLower(base)] = f

			short := strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(base, "spells-"), ".json"))
			spellFileMap[short] = f
		}

		for _, part := range parts {
			part = strings.ToLower(strings.TrimSpace(part))
			if part == "languages" || part == "lang" || part == "language" {
				shouldIngestLanguages = true
			} else if part == "conditions" || part == "cond" || part == "condition" {
				shouldIngestConditions = true
			} else if part == "spells" || part == "spell" {
				shouldIngestSpells = true
			} else {
				if f, ok := spellFileMap[part]; ok {
					targetSpellFiles = append(targetSpellFiles, f)
					shouldIngestSpells = true
				} else if f, ok := spellFileMap["spells-"+part+".json"]; ok {
					targetSpellFiles = append(targetSpellFiles, f)
					shouldIngestSpells = true
				} else {
					log.Fatalf("unknown ingestion target: %s", part)
				}
			}
		}
	} else {
		// Ingest everything by default
		shouldIngestLanguages = true
		shouldIngestConditions = true
		shouldIngestSpells = true
	}

	// Resolve spell files if spells should be ingested but none specified
	if shouldIngestSpells && len(targetSpellFiles) == 0 {
		var err error
		targetSpellFiles, err = getAvailableSpellFiles(*spellsDir)
		if err != nil {
			log.Fatalf("failed to scan spells: %v", err)
		}
	}

	ctx := context.Background()

	// 1. Ingest Languages
	if shouldIngestLanguages {
		langPath := filepath.Join(filepath.Dir(*spellsDir), "languages.json")
		if _, err := os.Stat(langPath); err == nil {
			fmt.Printf("Ingesting languages from %s...\n", langPath)
			nL, nS, err := db.IngestLanguagesFile(ctx, conn, langPath)
			if err != nil {
				log.Fatalf("languages ingestion: %v", err)
			}
			fmt.Printf("  - Ingested %d languages and %d scripts\n", nL, nS)
		} else {
			if *filesToIngest != "" {
				log.Fatalf("languages file not found at %s", langPath)
			}
		}
	}

	// 2. Ingest Conditions
	if shouldIngestConditions {
		condPath := filepath.Join(filepath.Dir(*spellsDir), "conditionsdiseases.json")
		if _, err := os.Stat(condPath); err == nil {
			fmt.Printf("Ingesting conditions from %s...\n", condPath)
			n, err := db.IngestConditionsFile(ctx, conn, condPath)
			if err != nil {
				log.Fatalf("conditions ingestion: %v", err)
			}
			fmt.Printf("  - Ingested %d conditions\n", n)
		} else {
			if *filesToIngest != "" {
				log.Fatalf("conditions file not found at %s", condPath)
			}
		}
	}

	// 3. Ingest Spells
	if shouldIngestSpells && len(targetSpellFiles) > 0 {
		total := 0
		for _, path := range targetSpellFiles {
			n, err := db.IngestFile(ctx, conn, path)
			if err != nil {
				log.Fatalf("%s: %v", path, err)
			}
			fmt.Printf("%s: %d spells ingested\n", filepath.Base(path), n)
			total += n
		}
		fmt.Printf("done, %d spells total ingested\n", total)
	}
}

func getAvailableSpellFiles(dir string) ([]string, error) {
	return filepath.Glob(filepath.Join(dir, "spells-*.json"))
}
