CREATE TABLE assets (
    id TEXT PRIMARY KEY,
    path TEXT NOT NULL,
    hash TEXT NOT NULL UNIQUE,
    mime_type TEXT,
    width INTEGER,
    height INTEGER,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE tags (
    id TEXT PRIMARY KEY,
    kind TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE tag_labels (
    tag_id TEXT NOT NULL,
    language TEXT NOT NULL,
    text TEXT NOT NULL,
    normalized TEXT NOT NULL,
    is_primary INTEGER NOT NULL DEFAULT 0,
    source TEXT NOT NULL,
    PRIMARY KEY (tag_id, language, normalized),
    FOREIGN KEY (tag_id) REFERENCES tags(id)
);

CREATE TABLE asset_tags (
    asset_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    source TEXT NOT NULL,
    confidence REAL,
    created_at INTEGER NOT NULL,
    PRIMARY KEY (asset_id, tag_id),
    FOREIGN KEY (asset_id) REFERENCES assets(id),
    FOREIGN KEY (tag_id) REFERENCES tags(id)
);

CREATE TABLE asset_perceptual_hashes (
    asset_id TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    hash_hex TEXT NOT NULL,
    bit_length INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    PRIMARY KEY (asset_id, algorithm),
    FOREIGN KEY (asset_id) REFERENCES assets(id)
);

CREATE INDEX idx_asset_tags_tag_id ON asset_tags(tag_id);
CREATE INDEX idx_asset_tags_asset_id ON asset_tags(asset_id);
CREATE INDEX idx_tag_labels_normalized ON tag_labels(normalized);
CREATE INDEX idx_asset_phash_algorithm_hash ON asset_perceptual_hashes (algorithm, hash_hex);
