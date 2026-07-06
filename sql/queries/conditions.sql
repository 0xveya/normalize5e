-- name: UpsertCondition :exec
INSERT INTO conditions (name, source, page, srd, basic_rules)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(name, source) DO UPDATE SET
    page = excluded.page,
    srd = excluded.srd,
    basic_rules = excluded.basic_rules;

-- name: DeleteConditionEntries :exec
DELETE FROM condition_entries
WHERE condition_name = ? AND source = ?;

-- name: InsertConditionEntry :exec
INSERT INTO condition_entries (condition_name, source, ord, block_type, heading, content)
VALUES (?, ?, ?, ?, ?, ?);
