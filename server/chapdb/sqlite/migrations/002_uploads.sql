-- +goose Up
CREATE TABLE uploads (
    id text primary key,
    size integer NOT NULL,
    uploader_id text NOT NULL,
    digests blob NOT NULL
);

-- +goose Down
DROP TABLE uploads;