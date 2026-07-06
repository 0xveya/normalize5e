package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"main/internal/sqlc"
)

type RawLanguage struct {
	Name            string          `json:"name"`
	Source          string          `json:"source"`
	Page            *int64          `json:"page"`
	Type            *string         `json:"type"`
	Script          *string         `json:"script"`
	Srd             json.RawMessage `json:"srd"`
	BasicRules      *bool           `json:"basicRules"`
	TypicalSpeakers []string        `json:"typicalSpeakers"`
}

type RawLanguageScript struct {
	Name   string   `json:"name"`
	Source string   `json:"source"`
	Fonts  []string `json:"fonts"`
}

type LanguagesFile struct {
	Language       []RawLanguage       `json:"language"`
	LanguageScript []RawLanguageScript `json:"languageScript"`
}

func IngestLanguagesFile(ctx context.Context, conn *sql.DB, path string) (int, int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}

	var lf LanguagesFile
	if err := json.Unmarshal(raw, &lf); err != nil {
		return 0, 0, fmt.Errorf("unmarshal: %w", err)
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	q := sqlc.New(tx)

	for _, lang := range lf.Language {
		srdBool, _ := parseSrd(lang.Srd)

		err := q.UpsertLanguage(ctx, sqlc.UpsertLanguageParams{
			Name:       lang.Name,
			Source:     lang.Source,
			Page:       toNullInt(lang.Page),
			Type:       toNullString(lang.Type),
			Script:     toNullString(lang.Script),
			Srd:        boolToNullInt(srdBool),
			BasicRules: boolToNullInt(deref(lang.BasicRules)),
		})
		if err != nil {
			return 0, 0, fmt.Errorf("upsert language %s: %w", lang.Name, err)
		}

		if err := q.DeleteLanguageSpeakers(ctx, sqlc.DeleteLanguageSpeakersParams{
			LanguageName: lang.Name,
			Source:       lang.Source,
		}); err != nil {
			return 0, 0, err
		}

		for _, speaker := range lang.TypicalSpeakers {
			err := q.InsertLanguageSpeaker(ctx, sqlc.InsertLanguageSpeakerParams{
				LanguageName: lang.Name,
				Source:       lang.Source,
				Speaker:      speaker,
			})
			if err != nil {
				return 0, 0, fmt.Errorf("insert language speaker %s: %w", speaker, err)
			}
		}
	}

	for _, script := range lf.LanguageScript {
		err := q.UpsertLanguageScript(ctx, sqlc.UpsertLanguageScriptParams{
			Name:   script.Name,
			Source: script.Source,
		})
		if err != nil {
			return 0, 0, fmt.Errorf("upsert script %s: %w", script.Name, err)
		}

		if err := q.DeleteLanguageScriptFonts(ctx, sqlc.DeleteLanguageScriptFontsParams{
			ScriptName: script.Name,
			Source:     script.Source,
		}); err != nil {
			return 0, 0, err
		}

		for _, font := range script.Fonts {
			err := q.InsertLanguageScriptFont(ctx, sqlc.InsertLanguageScriptFontParams{
				ScriptName: script.Name,
				Source:     script.Source,
				Font:       font,
			})
			if err != nil {
				return 0, 0, fmt.Errorf("insert script font %s: %w", font, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}

	return len(lf.Language), len(lf.LanguageScript), nil
}
