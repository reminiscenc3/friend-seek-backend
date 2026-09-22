CREATE TABLE IF NOT EXISTS users (
    login      TEXT PRIMARY KEY,
    note       TEXT     NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE IF NOT EXISTS locations (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    login      TEXT     NOT NULL REFERENCES users (login) ON DELETE CASCADE,
    latitude   REAL     NOT NULL,
    longitude  REAL     NOT NULL,
    azimuth    REAL     NOT NULL DEFAULT 0,
    accuracy   REAL     NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS locations_login_updated_at_idx
    ON locations (login, updated_at DESC, id DESC);
