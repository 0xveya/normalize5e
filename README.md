# normalize5e

Parses 5e JSON spell, language, and condition datasets into a SQLite database.

Created to assist [obsidian_dnd_vtt](https://github.com/mehrasmeydani/obsidian_dnd_vtt).

## Requirements

- Go
- sqlc (for code generation if modifying queries)

## Data Source

This utility requires the JSON datasets from the [5etools-src](https://github.com/5etools-mirror-3/5etools-src) repository to be located in the `data` directory (e.g., `data/languages.json`, `data/conditionsdiseases.json`, and the spells files within `data/spells/`).

## Building

```bash
go build -o normalize5e main.go
```

## CLI Usage

```text
  -db string
    	Path to SQLite database file (default: "data/db/dnd.db")
  -dir string
    	Directory containing spells-*.json files (default: "data/spells")
  -delete
    	Delete the SQLite database file and exit
  -list
    	List available ingestion targets and spell sources, then exit
  -files string
    	Comma-separated list of targets to ingest (e.g. "languages", "conditions", "spells", "phb", "tce")
```

### Examples

List sources:
```bash
./normalize5e -list
```

Ingest all files (spells, languages, and conditions):
```bash
./normalize5e
```

Ingest only languages and conditions:
```bash
./normalize5e -files languages,conditions
```

Ingest specific spell source files:
```bash
./normalize5e -files phb,tce
```

Delete database:
```bash
./normalize5e -delete
```

## Database Schema

Tables:
- `spells`: Main spell details, ranges, component flags, SRD properties.
- `spell_cast_time`: Casting times and conditions.
- `spell_duration` & `spell_duration_ends`: Spell durations, concentration, and end triggers.
- `spell_entries`: Text entries, tables, lists, and subentries.
- `spell_scaling_dice`: Damage scaling by spell level.
- `spell_tags`: Damage, saving throw, and condition tags.
- `languages`: Languages, types, and script associations.
- `language_typical_speakers`: List of creatures that typically speak each language.
- `language_scripts`: Written script systems.
- `language_script_fonts`: Display fonts associated with written script systems.
- `conditions`: Blinded, Charmed, Exhausted, etc., along with page references.
- `condition_entries`: Text effects, lists, and tables associated with each condition.

## Example Queries

### Find 3rd-level ritual spells:
```sql
SELECT name, source, school, level
FROM spells
WHERE level = 3 AND ritual = 1
ORDER BY name;
```

### Find concentration spells lasting >= 1 minute:
```sql
SELECT DISTINCT s.name, s.source, d.amount, d.time_unit
FROM spells s
JOIN spell_duration d ON d.spell_name = s.name AND d.source = s.source
WHERE d.concentration = 1 AND d.time_unit = 'minute'
ORDER BY d.amount DESC, s.name;
```

### Find all standard languages written in the Common script:
```sql
SELECT name, source, type
FROM languages
WHERE script = 'Common' AND type = 'standard'
ORDER BY name;
```

### Find the rules/text describing the Blinded condition:
```sql
SELECT ord, block_type, heading, content
FROM condition_entries
WHERE condition_name = 'Blinded'
ORDER BY ord;
```

## Verification & Data Comparison

To verify that the SQLite database accurately reflects the original JSON source files, you can use either Nushell or standard POSIX shell tools.

### Option 1: Using Nushell (Recommended)

Nushell supports reading SQLite databases directly as structured tables, making comparison scripts simple.

#### Verify Spell Ingestion:
Compare spell names from the `spells-aag.json` file directly against the SQLite database:
```nushell
let json_names = (open data/spells/spells-aag.json | get spell.name | sort)
let db_names = (open data/db/dnd.db | get spells | where source == "AAG" | get name | sort)
if $json_names == $db_names { echo "Verify: OK" } else { echo "Verify: MISMATCH" }
```

#### Query Database Tables:
Nushell lets you query, filter, and inspect database tables directly:
```nushell
open data/db/dnd.db | get spells | where LEVEL == 3 and ritual == 1 | select name source school
```

---

### Option 2: Using standard Shell (Bash + jq + sqlite3)

If you are not using Nushell, you can achieve the same check using standard command-line tools.

#### Verify Spell Ingestion:
```bash
# Extract names from JSON and sort
jq -r '.spell[].name' data/spells/spells-aag.json | sort > json_names.txt

# Extract names from SQLite database and sort
sqlite3 -list -noheader data/db/dnd.db "SELECT name FROM spells WHERE source = 'AAG' ORDER BY name;" > db_names.txt

# Compare the lists
diff json_names.txt db_names.txt && echo "Verify: OK" || echo "Verify: MISMATCH"

# Clean up
rm json_names.txt db_names.txt
```

