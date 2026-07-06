-- name: UpsertSpell :exec
INSERT INTO
    spells (
        name,
        source,
        LEVEL,
        school,
        PAGE,
        srd,
        srd_name,
        basic_rules,
        ritual,
        range_type,
        distance_type,
        distance_amount,
        comp_verbal,
        comp_somatic,
        comp_material,
        comp_material_text,
        comp_material_cost,
        comp_material_consumed
    )
VALUES
    (
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?
    ) ON CONFLICT(name, source) DO
UPDATE
SET
    LEVEL = excluded.level,
    school = excluded.school,
    PAGE = excluded.page,
    srd = excluded.srd,
    srd_name = excluded.srd_name,
    basic_rules = excluded.basic_rules,
    ritual = excluded.ritual,
    range_type = excluded.range_type,
    distance_type = excluded.distance_type,
    distance_amount = excluded.distance_amount,
    comp_verbal = excluded.comp_verbal,
    comp_somatic = excluded.comp_somatic,
    comp_material = excluded.comp_material,
    comp_material_text = excluded.comp_material_text,
    comp_material_cost = excluded.comp_material_cost,
    comp_material_consumed = excluded.comp_material_consumed;

-- name: DeleteCastTimes :exec
DELETE FROM
    spell_cast_time
WHERE
    spell_name = ?
    AND source = ?;

-- name: InsertCastTime :exec
INSERT INTO
    spell_cast_time (spell_name, source, ord, number, unit, condition)
VALUES
    (?, ?, ?, ?, ?, ?);

-- name: DeleteDurations :exec
DELETE FROM
    spell_duration
WHERE
    spell_name = ?
    AND source = ?;

-- name: InsertDuration :exec
INSERT INTO
    spell_duration (
        spell_name,
        source,
        ord,
        TYPE,
        amount,
        time_unit,
        concentration
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?);

-- name: DeleteDurationEnds :exec
DELETE FROM
    spell_duration_ends
WHERE
    spell_name = ?
    AND source = ?;

-- name: InsertDurationEnd :exec
INSERT INTO
    spell_duration_ends (spell_name, source, ord, ends_type)
VALUES
    (?, ?, ?, ?);

-- name: DeleteEntries :exec
DELETE FROM
    spell_entries
WHERE
    spell_name = ?
    AND source = ?;

-- name: InsertEntry :exec
INSERT INTO
    spell_entries (
        spell_name,
        source,
        section,
        ord,
        block_type,
        heading,
        content
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?);

-- name: DeleteScalingDice :exec
DELETE FROM
    spell_scaling_dice
WHERE
    spell_name = ?
    AND source = ?;

-- name: InsertScalingDice :exec
INSERT INTO
    spell_scaling_dice (spell_name, source, label, at_level, dice)
VALUES
    (?, ?, ?, ?, ?);

-- name: DeleteAltSources :exec
DELETE FROM
    spell_alt_sources
WHERE
    spell_name = ?
    AND source = ?;

-- name: InsertAltSource :exec
INSERT INTO
    spell_alt_sources (spell_name, source, kind, value)
VALUES
    (?, ?, ?, ?);

-- name: GetSpellsByLevel :many
SELECT
    name,
    LEVEL,
    school
FROM
    spells
WHERE
    LEVEL = ?
ORDER BY
    name;

-- name: GetSpellsBySavingThrow :many
SELECT
    s.name,
    s.source
FROM
    spells s
    JOIN spell_tags t ON t.spell_name = s.name
    AND t.source = s.source
WHERE
    t.tag_type = 'savingThrow'
    AND t.tag_value = ?;

-- name: DeleteSpellTags :exec
DELETE FROM
    spell_tags
WHERE
    spell_name = ?
    AND source = ?;

-- name: InsertSpellTag :exec
INSERT INTO
    spell_tags (spell_name, source, tag_type, tag_value)
VALUES
    (?, ?, ?, ?);

-- name: GetSpell :one
SELECT * FROM spells WHERE name = ? AND source = ?;

-- name: GetSpellCastTimes :many
SELECT * FROM spell_cast_time WHERE spell_name = ? AND source = ? ORDER BY ord;

-- name: GetSpellDurations :many
SELECT * FROM spell_duration WHERE spell_name = ? AND source = ? ORDER BY ord;

-- name: GetSpellDurationEnds :many
SELECT * FROM spell_duration_ends WHERE spell_name = ? AND source = ?;

-- name: GetSpellEntries :many
SELECT * FROM spell_entries WHERE spell_name = ? AND source = ? ORDER BY section, ord;

-- name: GetSpellScalingDice :many
SELECT * FROM spell_scaling_dice WHERE spell_name = ? AND source = ? ORDER BY at_level, label;

-- name: GetSpellTags :many
SELECT * FROM spell_tags WHERE spell_name = ? AND source = ? ORDER BY tag_type, tag_value;


