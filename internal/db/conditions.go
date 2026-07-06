package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"main/internal/sqlc"
)

type RawCondition struct {
	Name       string            `json:"name"`
	Source     string            `json:"source"`
	Page       *int64            `json:"page"`
	Srd        json.RawMessage   `json:"srd"`
	BasicRules *bool             `json:"basicRules"`
	Entries    []json.RawMessage `json:"entries"`
}

type ConditionsFile struct {
	Condition []RawCondition `json:"condition"`
}

func IngestConditionsFile(ctx context.Context, conn *sql.DB, path string) (int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	var cf ConditionsFile
	if err := json.Unmarshal(raw, &cf); err != nil {
		return 0, fmt.Errorf("unmarshal: %w", err)
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	q := sqlc.New(tx)

	for _, cond := range cf.Condition {
		srdBool, _ := parseSrd(cond.Srd)

		err := q.UpsertCondition(ctx, sqlc.UpsertConditionParams{
			Name:       cond.Name,
			Source:     cond.Source,
			Page:       toNullInt(cond.Page),
			Srd:        boolToNullInt(srdBool),
			BasicRules: boolToNullInt(deref(cond.BasicRules)),
		})
		if err != nil {
			return 0, fmt.Errorf("upsert condition %s: %w", cond.Name, err)
		}

		if err := q.DeleteConditionEntries(ctx, sqlc.DeleteConditionEntriesParams{
			ConditionName: cond.Name,
			Source:        cond.Source,
		}); err != nil {
			return 0, err
		}

		if err := insertConditionEntryBlocks(ctx, q, cond.Name, cond.Source, cond.Entries); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return len(cf.Condition), nil
}

func insertConditionEntryBlocks(ctx context.Context, q *sqlc.Queries, name, source string, entries []json.RawMessage) error {
	ord := int64(0)
	walk := func(items []json.RawMessage) error {
		for _, raw := range items {
			var asString string
			if err := json.Unmarshal(raw, &asString); err == nil {
				if err := q.InsertConditionEntry(ctx, sqlc.InsertConditionEntryParams{
					ConditionName: name,
					Source:        source,
					Ord:           ord,
					BlockType:     "text",
					Heading:       sql.NullString{Valid: false},
					Content:       asString,
				}); err != nil {
					return err
				}
				ord++
				continue
			}

			var obj struct {
				Type    string            `json:"type"`
				Name    string            `json:"name"`
				Entries []json.RawMessage `json:"entries"`
			}
			if err := json.Unmarshal(raw, &obj); err != nil {
				return fmt.Errorf("unrecognized entry shape: %w", err)
			}

			switch obj.Type {
			case "entries":
				if err := q.InsertConditionEntry(ctx, sqlc.InsertConditionEntryParams{
					ConditionName: name,
					Source:        source,
					Ord:           ord,
					BlockType:     "subentry",
					Heading:       strToNullString(obj.Name),
					Content:       string(raw),
				}); err != nil {
					return err
				}
				ord++
			default:
				blockType := obj.Type
				if blockType == "" {
					blockType = "unknown"
				}
				if err := q.InsertConditionEntry(ctx, sqlc.InsertConditionEntryParams{
					ConditionName: name,
					Source:        source,
					Ord:           ord,
					BlockType:     blockType,
					Heading:       sql.NullString{Valid: false},
					Content:       string(raw),
				}); err != nil {
					return err
				}
				ord++
			}
		}
		return nil
	}
	return walk(entries)
}
