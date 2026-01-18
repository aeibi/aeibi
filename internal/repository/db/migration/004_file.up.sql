CREATE TABLE files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size INTEGER NOT NULL,
    checksum TEXT NOT NULL,
    uploader TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('NORMAL', 'ARCHIVED')) DEFAULT 'NORMAL',
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE UNIQUE INDEX idx_file_url ON files (url);
CREATE INDEX idx_file_uploader ON files (uploader);
CREATE INDEX idx_file_checksum ON files (checksum);
