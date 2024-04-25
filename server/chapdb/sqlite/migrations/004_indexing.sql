-- +goose Up
ALTER TABLE objects ADD COLUMN indexed_at DATETIME NOT NULL;

CREATE TABLE object_errors (
    id INTEGER PRIMARY KEY, -- internal db ID
    store_id TEXT NOT NULL, -- storage root ID
    path TEXT NOT NULL, -- path to object with error (relative to FS)
    ocfl_id TEXT NOT NULL, -- object's ocfl_id (may be empty string)
    ocfl_id_missing BOOLEAN NOT NULL, -- boolean indicating missing object ID
    error TEXT NOT NULL, -- error message text
    reported_at DATETIME NOT NULL -- date when error was reported
);

-- +goose Down
ALTER TABLE objects DROP COLUMN indexed_at;
DROP TABLE object_errors;
