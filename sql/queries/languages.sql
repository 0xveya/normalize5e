-- name: UpsertLanguage :exec
INSERT INTO languages (name, source, page, type, script, srd, basic_rules)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(name, source) DO UPDATE SET
    page = excluded.page,
    type = excluded.type,
    script = excluded.script,
    srd = excluded.srd,
    basic_rules = excluded.basic_rules;

-- name: DeleteLanguageSpeakers :exec
DELETE FROM language_typical_speakers
WHERE language_name = ? AND source = ?;

-- name: InsertLanguageSpeaker :exec
INSERT INTO language_typical_speakers (language_name, source, speaker)
VALUES (?, ?, ?);

-- name: UpsertLanguageScript :exec
INSERT INTO language_scripts (name, source)
VALUES (?, ?)
ON CONFLICT(name, source) DO NOTHING;

-- name: DeleteLanguageScriptFonts :exec
DELETE FROM language_script_fonts
WHERE script_name = ? AND source = ?;

-- name: InsertLanguageScriptFont :exec
INSERT INTO language_script_fonts (script_name, source, font)
VALUES (?, ?, ?);
