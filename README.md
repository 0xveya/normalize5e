# normalize5e

Parses 5e JSON spell datasets into a SQLite database. 

Created to assist [obsidian_dnd_vtt](https://github.com/mehrasmeydani/obsidian_dnd_vtt).

## Requirements

- Go
- sqlc (for code generation if modifying queries)

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
    	List available spell files/sources in the directory and exit
  -files string
    	Comma-separated list of spell sources/filenames to ingest (e.g. "phb,tce")
```

### Examples

List sources:
```bash
./normalize5e -list
```

Ingest all files:
```bash
./normalize5e
```

Ingest specific files:
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

## Example Queries

### Find 3rd-level ritual spells:
```sql
SELECT name, source, school, level
FROM spells
WHERE level = 3 AND ritual = 1
ORDER BY name;
```

### Find spells inflicting fire damage:
```sql
SELECT DISTINCT s.name, s.source, s.level
FROM spells s
JOIN spell_tags t ON t.spell_name = s.name AND t.source = s.source
WHERE t.tag_type = 'damageInflict' AND t.tag_value = 'fire'
ORDER BY s.level, s.name;
```

### Find concentration spells lasting >= 1 minute:
```sql
SELECT DISTINCT s.name, s.source, d.amount, d.time_unit
FROM spells s
JOIN spell_duration d ON d.spell_name = s.name AND d.source = s.source
WHERE d.concentration = 1 AND d.time_unit = 'minute'
ORDER BY d.amount DESC, s.name;
```

### Find spells with Verbal & Somatic components but no Material components:
```sql
SELECT name, source, level
FROM spells
WHERE comp_verbal = 1 AND comp_somatic = 1 AND comp_material = 0
ORDER BY level, name;
```

### Find level-scaling dice for Booming Blade:
```sql
SELECT at_level, label, dice
FROM spell_scaling_dice
WHERE spell_name = 'Booming Blade'
ORDER BY at_level;
```
