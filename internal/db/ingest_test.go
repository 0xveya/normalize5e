package db

import (
	"context"
	"os"
	"testing"

	"main/internal/sqlc"
)

func TestIngestion(t *testing.T) {
	dbPath := "test.db"
	_ = os.Remove(dbPath)
	defer os.Remove(dbPath)

	schemaBytes, err := os.ReadFile("../../sql/schema.sql")
	if err != nil {
		t.Fatalf("failed to read schema: %v", err)
	}

	conn, err := Connect(dbPath, string(schemaBytes))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	// Ingest spells
	spellCount, err := IngestFile(ctx, conn, "../../data/spells/spells-aag.json")
	if err != nil {
		t.Fatalf("failed to ingest spells: %v", err)
	}
	if spellCount != 2 {
		t.Errorf("expected 2 spells, got %d", spellCount)
	}

	// Ingest languages
	langCount, scriptCount, err := IngestLanguagesFile(ctx, conn, "../../data/languages.json")
	if err != nil {
		t.Fatalf("failed to ingest languages: %v", err)
	}
	if langCount == 0 || scriptCount == 0 {
		t.Errorf("expected non-zero languages/scripts, got %d/%d", langCount, scriptCount)
	}

	// Ingest conditions
	condCount, err := IngestConditionsFile(ctx, conn, "../../data/conditionsdiseases.json")
	if err != nil {
		t.Fatalf("failed to ingest conditions: %v", err)
	}
	if condCount == 0 {
		t.Errorf("expected non-zero conditions, got %d", condCount)
	}

	q := sqlc.New(conn)

	// Verify "Air Bubble" spell details
	spell, err := q.GetSpell(ctx, sqlc.GetSpellParams{Name: "Air Bubble", Source: "AAG"})
	if err != nil {
		t.Fatalf("failed to get spell: %v", err)
	}
	if spell.Level.Int64 != 2 {
		t.Errorf("expected level 2, got %d", spell.Level.Int64)
	}
	if spell.School.String != "C" {
		t.Errorf("expected school C, got %s", spell.School.String)
	}
	if spell.Page.Int64 != 22 {
		t.Errorf("expected page 22, got %d", spell.Page.Int64)
	}
	if spell.CompSomatic.Int64 != 1 || spell.CompVerbal.Int64 != 0 || spell.CompMaterial.Int64 != 0 {
		t.Errorf("incorrect components: somatic=%d, verbal=%d, material=%d", spell.CompSomatic.Int64, spell.CompVerbal.Int64, spell.CompMaterial.Int64)
	}

	// Verify "Air Bubble" cast times
	castTimes, err := q.GetSpellCastTimes(ctx, sqlc.GetSpellCastTimesParams{SpellName: "Air Bubble", Source: "AAG"})
	if err != nil {
		t.Fatalf("failed to get cast times: %v", err)
	}
	if len(castTimes) != 1 || castTimes[0].Number.Int64 != 1 || castTimes[0].Unit.String != "action" {
		t.Errorf("incorrect cast times: %+v", castTimes)
	}

	// Verify "Air Bubble" duration
	durations, err := q.GetSpellDurations(ctx, sqlc.GetSpellDurationsParams{SpellName: "Air Bubble", Source: "AAG"})
	if err != nil {
		t.Fatalf("failed to get durations: %v", err)
	}
	if len(durations) != 1 || durations[0].Amount.Int64 != 24 || durations[0].TimeUnit.String != "hour" {
		t.Errorf("incorrect durations: %+v", durations)
	}

	// Verify "Create Spelljamming Helm" material components
	helmSpell, err := q.GetSpell(ctx, sqlc.GetSpellParams{Name: "Create Spelljamming Helm", Source: "AAG"})
	if err != nil {
		t.Fatalf("failed to get helm spell: %v", err)
	}
	if helmSpell.CompMaterialCost.Int64 != 500000 || helmSpell.CompMaterialConsumed.Int64 != 1 {
		t.Errorf("incorrect material components: cost=%d, consumed=%d", helmSpell.CompMaterialCost.Int64, helmSpell.CompMaterialConsumed.Int64)
	}

	// Verify language "Abyssal"
	lang, err := q.GetLanguage(ctx, sqlc.GetLanguageParams{Name: "Abyssal", Source: "PHB"})
	if err != nil {
		t.Fatalf("failed to get language Abyssal: %v", err)
	}
	if lang.Script.String != "Infernal" {
		t.Errorf("expected script Infernal, got %s", lang.Script.String)
	}

	speakers, err := q.GetLanguageSpeakers(ctx, sqlc.GetLanguageSpeakersParams{LanguageName: "Abyssal", Source: "PHB"})
	if err != nil {
		t.Fatalf("failed to get language speakers: %v", err)
	}
	foundDemonSpeaker := false
	for _, s := range speakers {
		if s == "{@filter demons|bestiary|tag=demon}" {
			foundDemonSpeaker = true
			break
		}
	}
	if !foundDemonSpeaker {
		t.Errorf("did not find demon speaker in: %+v", speakers)
	}

	// Verify condition "Blinded"
	cond, err := q.GetCondition(ctx, sqlc.GetConditionParams{Name: "Blinded", Source: "PHB"})
	if err != nil {
		t.Fatalf("failed to get condition: %v", err)
	}
	if cond.Page.Int64 != 290 {
		t.Errorf("expected page 290, got %d", cond.Page.Int64)
	}
}
