CREATE TABLE IF NOT EXISTS spells (
    name TEXT NOT NULL,
    source TEXT NOT NULL,
    LEVEL INTEGER,
    school TEXT,
    PAGE INTEGER,
    srd INTEGER DEFAULT 0,
    srd_name TEXT,
    basic_rules INTEGER DEFAULT 0,
    ritual INTEGER DEFAULT 0,
    range_type TEXT,
    distance_type TEXT,
    distance_amount INTEGER,
    comp_verbal INTEGER DEFAULT 0,
    comp_somatic INTEGER DEFAULT 0,
    comp_material INTEGER DEFAULT 0,
    comp_material_text TEXT,
    comp_material_cost INTEGER,
    comp_material_consumed INTEGER DEFAULT 0,
    PRIMARY KEY (name, source)
);

CREATE TABLE IF NOT EXISTS spell_cast_time (
    spell_name TEXT NOT NULL,
    source TEXT NOT NULL,
    ord INTEGER NOT NULL,
    number INTEGER,
    unit TEXT,
    condition TEXT,
    PRIMARY KEY (spell_name, source, ord)
);

CREATE TABLE IF NOT EXISTS spell_duration (
    spell_name TEXT NOT NULL,
    source TEXT NOT NULL,
    ord INTEGER NOT NULL,
    TYPE TEXT,
    amount INTEGER,
    time_unit TEXT,
    concentration INTEGER DEFAULT 0,
    PRIMARY KEY (spell_name, source, ord)
);

CREATE TABLE IF NOT EXISTS spell_duration_ends (
    spell_name TEXT NOT NULL,
    source TEXT NOT NULL,
    ord INTEGER NOT NULL,
    ends_type TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS spell_entries (
    spell_name TEXT NOT NULL,
    source TEXT NOT NULL,
    section TEXT NOT NULL DEFAULT 'main',
    ord INTEGER NOT NULL,
    block_type TEXT NOT NULL,
    heading TEXT,
    content TEXT NOT NULL,
    PRIMARY KEY (spell_name, source, section, ord)
);

CREATE TABLE IF NOT EXISTS spell_scaling_dice (
    spell_name TEXT NOT NULL,
    source TEXT NOT NULL,
    label TEXT NOT NULL DEFAULT '',
    at_level INTEGER NOT NULL,
    dice TEXT NOT NULL,
    PRIMARY KEY (spell_name, source, at_level, label)
);

CREATE TABLE IF NOT EXISTS spell_alt_sources (
    spell_name TEXT NOT NULL,
    source TEXT NOT NULL,
    kind TEXT NOT NULL,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS spell_tags (
    spell_name TEXT NOT NULL,
    source TEXT NOT NULL,
    tag_type TEXT NOT NULL,
    tag_value TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_spell_tags_lookup ON spell_tags(tag_type, tag_value);

-- Languages
CREATE TABLE IF NOT EXISTS languages (
    name TEXT NOT NULL,
    source TEXT NOT NULL,
    page INTEGER,
    type TEXT,
    script TEXT,
    srd INTEGER DEFAULT 0,
    basic_rules INTEGER DEFAULT 0,
    PRIMARY KEY (name, source)
);

CREATE TABLE IF NOT EXISTS language_typical_speakers (
    language_name TEXT NOT NULL,
    source TEXT NOT NULL,
    speaker TEXT NOT NULL,
    PRIMARY KEY (language_name, source, speaker)
);

CREATE TABLE IF NOT EXISTS language_scripts (
    name TEXT NOT NULL,
    source TEXT NOT NULL,
    PRIMARY KEY (name, source)
);

CREATE TABLE IF NOT EXISTS language_script_fonts (
    script_name TEXT NOT NULL,
    source TEXT NOT NULL,
    font TEXT NOT NULL,
    PRIMARY KEY (script_name, source, font)
);

-- Conditions
CREATE TABLE IF NOT EXISTS conditions (
    name TEXT NOT NULL,
    source TEXT NOT NULL,
    page INTEGER,
    srd INTEGER DEFAULT 0,
    basic_rules INTEGER DEFAULT 0,
    PRIMARY KEY (name, source)
);

CREATE TABLE IF NOT EXISTS condition_entries (
    condition_name TEXT NOT NULL,
    source TEXT NOT NULL,
    ord INTEGER NOT NULL,
    block_type TEXT NOT NULL,
    heading TEXT,
    content TEXT NOT NULL,
    PRIMARY KEY (condition_name, source, ord)
);

