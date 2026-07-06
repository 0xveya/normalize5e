package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"main/internal/sqlc"
)

// ---------- raw JSON shapes (only the messy nested fields) ----------

type RawSpell struct {
	Name       string          `json:"name"`
	Source     string          `json:"source"`
	Level      *int64          `json:"level"`
	School     *string         `json:"school"`
	Page       *int64          `json:"page"`
	Srd        json.RawMessage `json:"srd"`
	BasicRules *bool           `json:"basicRules"`

	Time []struct {
		Number    *int64  `json:"number"`
		Unit      *string `json:"unit"`
		Condition *string `json:"condition"`
	} `json:"time"`

	Range struct {
		Type     string `json:"type"`
		Distance struct {
			Type   string `json:"type"`
			Amount *int64 `json:"amount"`
		} `json:"distance"`
	} `json:"range"`

	Components struct {
		V bool            `json:"v"`
		S bool            `json:"s"`
		M json.RawMessage `json:"m"`
	} `json:"components"`

	Duration []struct {
		Type     string `json:"type"`
		Duration *struct {
			Type   string `json:"type"`
			Amount *int64 `json:"amount"`
		} `json:"duration"`
		Concentration *bool    `json:"concentration"`
		Ends          []string `json:"ends"`
	} `json:"duration"`

	Entries            []json.RawMessage `json:"entries"`
	EntriesHigherLevel []struct {
		Name    string            `json:"name"`
		Entries []json.RawMessage `json:"entries"`
	} `json:"entriesHigherLevel"`

	ScalingLevelDice *struct {
		Label   string            `json:"label"`
		Scaling map[string]string `json:"scaling"`
	} `json:"scalingLevelDice"`

	Meta struct {
		Ritual *bool `json:"ritual"`
	} `json:"meta"`

	DamageInflict       []string `json:"damageInflict"`
	SavingThrow         []string `json:"savingThrow"`
	ConditionInflict    []string `json:"conditionInflict"`
	ConditionImmune     []string `json:"conditionImmune"`
	DamageResist        []string `json:"damageResist"`
	DamageImmune        []string `json:"damageImmune"`
	DamageVulnerable    []string `json:"damageVulnerable"`
	AffectsCreatureType []string `json:"affectsCreatureType"`
	MiscTags            []string `json:"miscTags"`
	AreaTags            []string `json:"areaTags"`
	SpellAttack         []string `json:"spellAttack"`
	AbilityCheck        []string `json:"abilityCheck"`
}

type SpellFile struct {
	Spell []RawSpell `json:"spell"`
}

// IngestFile reads a spell JSON file and inserts/upserts all spells within it using a transaction.
func IngestFile(ctx context.Context, conn *sql.DB, path string) (int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	var sf SpellFile
	if err := json.Unmarshal(raw, &sf); err != nil {
		return 0, fmt.Errorf("unmarshal: %w", err)
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	q := sqlc.New(tx)

	for _, sp := range sf.Spell {
		if err := InsertSpell(ctx, q, sp); err != nil {
			return 0, fmt.Errorf("insert %s: %w", sp.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(sf.Spell), nil
}

// InsertSpell flattens and inserts one spell using sqlc.
func InsertSpell(ctx context.Context, q *sqlc.Queries, sp RawSpell) error {
	name, source := sp.Name, sp.Source

	srdBool, srdName := parseSrd(sp.Srd)
	matText, matCost, matConsumed := parseMaterial(sp.Components.M)

	err := q.UpsertSpell(ctx, sqlc.UpsertSpellParams{
		Name:                 name,
		Source:               source,
		Level:                toNullInt(sp.Level),
		School:               toNullString(sp.School),
		Page:                 toNullInt(sp.Page),
		Srd:                  boolToNullInt(srdBool),
		SrdName:              toNullString(srdName),
		BasicRules:           boolToNullInt(deref(sp.BasicRules)),
		Ritual:               boolToNullInt(deref(sp.Meta.Ritual)),
		RangeType:            strToNullString(sp.Range.Type),
		DistanceType:         strToNullString(sp.Range.Distance.Type),
		DistanceAmount:       toNullInt(sp.Range.Distance.Amount),
		CompVerbal:           boolToNullInt(sp.Components.V),
		CompSomatic:          boolToNullInt(sp.Components.S),
		CompMaterial:         boolToNullInt(matText != nil),
		CompMaterialText:     toNullString(matText),
		CompMaterialCost:     toNullInt(matCost),
		CompMaterialConsumed: boolToNullInt(matConsumed),
	})
	if err != nil {
		return fmt.Errorf("upsert spells row: %w", err)
	}

	// --- cast times ---
	if err := q.DeleteCastTimes(ctx, sqlc.DeleteCastTimesParams{SpellName: name, Source: source}); err != nil {
		return err
	}
	for i, ct := range sp.Time {
		if err := q.InsertCastTime(ctx, sqlc.InsertCastTimeParams{
			SpellName: name,
			Source:    source,
			Ord:       int64(i),
			Number:    toNullInt(ct.Number),
			Unit:      toNullString(ct.Unit),
			Condition: toNullString(ct.Condition),
		}); err != nil {
			return fmt.Errorf("cast time: %w", err)
		}
	}

	// --- durations + ends ---
	if err := q.DeleteDurations(ctx, sqlc.DeleteDurationsParams{SpellName: name, Source: source}); err != nil {
		return err
	}
	if err := q.DeleteDurationEnds(ctx, sqlc.DeleteDurationEndsParams{SpellName: name, Source: source}); err != nil {
		return err
	}
	for i, d := range sp.Duration {
		var amount *int64
		var timeUnit *string
		if d.Duration != nil {
			amount = d.Duration.Amount
			timeUnit = nullableStr(d.Duration.Type)
		}
		if err := q.InsertDuration(ctx, sqlc.InsertDurationParams{
			SpellName:     name,
			Source:        source,
			Ord:           int64(i),
			Type:          strToNullString(d.Type),
			Amount:        toNullInt(amount),
			TimeUnit:      toNullString(timeUnit),
			Concentration: boolToNullInt(deref(d.Concentration)),
		}); err != nil {
			return fmt.Errorf("duration: %w", err)
		}
		for _, end := range d.Ends {
			if err := q.InsertDurationEnd(ctx, sqlc.InsertDurationEndParams{
				SpellName: name,
				Source:    source,
				Ord:       int64(i),
				EndsType:  end,
			}); err != nil {
				return fmt.Errorf("duration end: %w", err)
			}
		}
	}

	// --- entries (main + higher level), flattened one level deep ---
	if err := q.DeleteEntries(ctx, sqlc.DeleteEntriesParams{SpellName: name, Source: source}); err != nil {
		return err
	}
	if err := insertEntryBlocks(ctx, q, name, source, "main", sp.Entries); err != nil {
		return err
	}
	for _, hl := range sp.EntriesHigherLevel {
		if err := insertEntryBlocks(ctx, q, name, source, "higher_level", hl.Entries); err != nil {
			return err
		}
	}

	// --- scaling dice ---
	if err := q.DeleteScalingDice(ctx, sqlc.DeleteScalingDiceParams{SpellName: name, Source: source}); err != nil {
		return err
	}
	if sp.ScalingLevelDice != nil {
		for lvl, dice := range sp.ScalingLevelDice.Scaling {
			var lvlInt int64
			fmt.Sscanf(lvl, "%d", &lvlInt)
			if err := q.InsertScalingDice(ctx, sqlc.InsertScalingDiceParams{
				SpellName: name,
				Source:    source,
				Label:     strToNullString(sp.ScalingLevelDice.Label),
				AtLevel:   lvlInt,
				Dice:      dice,
			}); err != nil {
				return fmt.Errorf("scaling dice: %w", err)
			}
		}
	}

	// --- tag fields ---
	if err := q.DeleteSpellTags(ctx, sqlc.DeleteSpellTagsParams{SpellName: name, Source: source}); err != nil {
		return err
	}
	tagGroups := map[string][]string{
		"damageInflict":       sp.DamageInflict,
		"savingThrow":         sp.SavingThrow,
		"conditionInflict":    sp.ConditionInflict,
		"conditionImmune":     sp.ConditionImmune,
		"damageResist":        sp.DamageResist,
		"damageImmune":        sp.DamageImmune,
		"damageVulnerable":    sp.DamageVulnerable,
		"affectsCreatureType": sp.AffectsCreatureType,
		"miscTags":            sp.MiscTags,
		"areaTags":            sp.AreaTags,
		"spellAttack":         sp.SpellAttack,
		"abilityCheck":        sp.AbilityCheck,
	}
	for tagType, values := range tagGroups {
		for _, v := range values {
			if err := q.InsertSpellTag(ctx, sqlc.InsertSpellTagParams{
				SpellName: name,
				Source:    source,
				TagType:   tagType,
				TagValue:  v,
			}); err != nil {
				return fmt.Errorf("tag %s: %w", tagType, err)
			}
		}
	}

	return nil
}

// insertEntryBlocks flattens one level of the entries[] tree.
func insertEntryBlocks(ctx context.Context, q *sqlc.Queries, name, source, section string, entries []json.RawMessage) error {
	ord := int64(0)
	var walk func(items []json.RawMessage) error
	walk = func(items []json.RawMessage) error {
		for _, raw := range items {
			var asString string
			if err := json.Unmarshal(raw, &asString); err == nil {
				if err := q.InsertEntry(ctx, sqlc.InsertEntryParams{
					SpellName: name,
					Source:    source,
					Section:   section,
					Ord:       ord,
					BlockType: "text",
					Heading:   sql.NullString{Valid: false},
					Content:   asString,
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
				if err := q.InsertEntry(ctx, sqlc.InsertEntryParams{
					SpellName: name,
					Source:    source,
					Section:   section,
					Ord:       ord,
					BlockType: "subentry",
					Heading:   strToNullString(obj.Name),
					Content:   string(raw),
				}); err != nil {
					return err
				}
				ord++
			default: // "list", "table", or anything else — store the raw block as-is
				blockType := obj.Type
				if blockType == "" {
					blockType = "unknown"
				}
				if err := q.InsertEntry(ctx, sqlc.InsertEntryParams{
					SpellName: name,
					Source:    source,
					Section:   section,
					Ord:       ord,
					BlockType: blockType,
					Heading:   sql.NullString{Valid: false},
					Content:   string(raw),
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

// ---------- helper functions ----------

func toNullInt(val *int64) sql.NullInt64 {
	if val == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *val, Valid: true}
}

func boolToNullInt(b bool) sql.NullInt64 {
	val := int64(0)
	if b {
		val = 1
	}
	return sql.NullInt64{Int64: val, Valid: true}
}

func toNullString(val *string) sql.NullString {
	if val == nil || *val == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *val, Valid: true}
}

func strToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func deref(b *bool) bool {
	return b != nil && *b
}

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// parseSrd handles srd being: absent, true/false, or a string (renamed SRD name).
func parseSrd(raw json.RawMessage) (bool, *string) {
	if len(raw) == 0 || string(raw) == "null" {
		return false, nil
	}
	var b bool
	if json.Unmarshal(raw, &b) == nil {
		return b, nil
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return true, &s
	}
	return false, nil
}

// parseMaterial handles components.m being: absent, a plain string, or an object {text, cost, consume}.
func parseMaterial(raw json.RawMessage) (matText *string, matCost *int64, consumed bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil, false
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return &s, nil, false
	}
	var obj struct {
		Text    string `json:"text"`
		Cost    *int64 `json:"cost"`
		Consume *bool  `json:"consume"`
	}
	if json.Unmarshal(raw, &obj) == nil {
		return &obj.Text, obj.Cost, deref(obj.Consume)
	}
	return nil, nil, false
}
